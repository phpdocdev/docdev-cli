package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-ping/ping"
	"github.com/jedib0t/go-pretty/v6/table"
	txt "github.com/jedib0t/go-pretty/v6/text"
	"github.com/joho/godotenv"
	"github.com/txn2/txeh"
	"github.com/urfave/cli/v2"
)

var Version = "development"

func main() {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		path := os.Getenv("DOCDEV_PATH")
		os.Chdir(path)
	}
	loadEnv()

	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"d"},
			Usage:   "Dry run",
		},
	}

	app := &cli.App{
		EnableBashCompletion: true,
		Flags:                flags,
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize configuration and install mkcert",
				Action:  Init,
				Flags: append([]cli.Flag{
					&cli.StringFlag{
						Name:    "tld",
						Aliases: []string{"t"},
						Value:   "loc",
						Usage:   "TLD for project hostnames",
					},
					&cli.StringFlag{
						Name:    "root",
						Aliases: []string{"r"},
						Value:   os.Getenv("HOME") + "/repos/",
						Usage:   "Root directory containing your projects",
					},
					&cli.StringFlag{
						Name:    "php",
						Aliases: []string{"p"},
						Value:   "74",
						Usage:   "Initial PHP version",
					},
					&cli.BoolFlag{
						Name:  "certs",
						Usage: "Generate and install certificates",
					},
					&cli.BoolFlag{
						Name:  "hosts",
						Usage: "Generate hosts file",
					},
					&cli.BoolFlag{
						Name:  "start",
						Usage: "Start containers immediately",
					},
				}, flags...),
			},
			{
				Name:    "certs",
				Aliases: []string{"c"},
				Usage:   "Generate and install the certificates",
				Action:  GenerateCerts,
			},
			{
				Name:    "hosts",
				Aliases: []string{},
				Usage:   "Generate a new hosts profile and add it to your system /etc/host",
				Action:  GenerateHosts,
				Flags:   flags,
			},
			{
				Name:    "start",
				Aliases: []string{"s"},
				Usage:   "Bring up the docker containers",
				Action:  StartContainer,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "php-only",
						Usage: "Reset the PHP container",
						Value: true,
					},
					&cli.BoolFlag{
						Name:    "exec",
						Aliases: []string{"e"},
						Usage:   "Start container shell after starting",
					},
				},
			},
			{
				Name:    "exec",
				Aliases: []string{"e"},
				Usage:   "Start docker container shell",
				Action:  ExecContainer,
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Test your configuration.",
				Action:  TestConfiguration,
			},
			{
				Name:    "php",
				Aliases: []string{"p"},
				Usage:   "Change php version (requires \"start\" to rebuild). Valid values: 54, 56, 72, 74",
				Action:  ChangePhpVersion,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "start",
						Aliases: []string{"s"},
						Usage:   "Start the containers after switching the PHP version",
					},
				},
			},
			{
				Name:   "refresh",
				Usage:  "Pull changes from git and images from Docker",
				Action: Refresh,
			},
			{
				Name:   "selfupdate",
				Usage:  "Update docdev binary",
				Action: SelfUpdate,
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Output current version.",
				Action:  PrintVersion,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func PrintVersion(c *cli.Context) error {
	fmt.Println(Version)
	return nil
}

func Refresh(c *cli.Context) error {
	fmt.Println("Removing containers and images")
	cmd := "docker-compose down --rmi all"
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Pulling changes from Git")
	cmd = "git pull origin master"
	_, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Restarting...")
	return StartContainer(c)
}

func SelfUpdate(c *cli.Context) error {

	fmt.Println("Downloading latest release from github")

	release := fmt.Sprintf("docdev-%s-%s", runtime.GOOS, runtime.GOARCH)
	cmd := fmt.Sprintf("gh release download -p \"%s\" --repo \"https://github.ark.org/brandon-kiefer/docker-dev\"", release)
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	fmt.Println("Attempting to replace the existing binary")
	os.Rename(release, "docdev")
	newpath, err := exec.Command("which", "docdev").Output()
	exists := strings.Replace(string(newpath), "\n", "", -1)
	if exists != "" {
		input, err := ioutil.ReadFile("docdev")
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = ioutil.WriteFile(exists, input, 0644)
		if err != nil {
			fmt.Println("Error creating", exists)
			fmt.Println(err)
			return err
		}
	}

	fmt.Println("docdev has been updated!")
	return nil
}

func Init(c *cli.Context) error {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		_, err := exec.Command("cp", ".env.example", ".env").Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
		setEnvFileValue("TLD_SUFFIX", c.String("tld"))
		setEnvFileValue("DOCUMENTROOT", c.String("root"))
		setEnvFileValue("PHPV", c.String("php"))

		path, _ := os.Getwd()
		setRcExport("DOCDEV_PATH", path)

		fmt.Printf("%s", "Created .env file\n")
	} else {
		fmt.Printf("%s", ".env file already exists.\n")
	}

	mkcert, err := exec.Command("which", "mkcert").Output()

	if string(mkcert[:]) == "" {
		_, err = exec.Command("brew", "install", "mkcert").Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
	}

	if c.Bool("certs") {
		fmt.Printf("%s", "Generating certificates...\n")
		err = GenerateCerts(c)
		if err != nil {
			return cli.Exit(err, 86)
		}
	}
	if c.Bool("hosts") {
		fmt.Printf("%s", "Generating hosts...\n")
		err = GenerateHosts(c)
		if err != nil {
			return cli.Exit(err, 86)
		}
	}
	if c.Bool("start") {
		fmt.Printf("%s", "Starting containers...\n")
		err = StartContainer(c)
		if err != nil {
			return cli.Exit(err, 86)
		}
	}

	return err
}

func TestConfiguration(c *cli.Context) error {

	tw := table.NewWriter()
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	tw.SetOutputMirror(os.Stdout)

	phpv := os.Getenv("DOCDEV_PHP")
	ddPath := os.Getenv("DOCDEV_PATH")
	tw.AppendRows([]table.Row{
		{"$USER DOCDEV_PHP: ", phpv},
		{"$USER DOCDEV_PATH: ", ddPath},
	})
	tw.AppendRow(table.Row{"---"})

	mkcert, _ := exec.Command("which", "mkcert").Output()
	hostctl, _ := exec.Command("which", "hostctl").Output()
	tw.AppendRows([]table.Row{
		{"mkcert installed: ", strings.Replace(string(mkcert), "\n", "", -1)},
		{"hostctl installed: ", strings.Replace(string(hostctl), "\n", "", -1)},
	}, rowConfigAutoMerge)
	tw.AppendRow(table.Row{"---"})

	certInstalled := ""
	if isCertInstalled() != "" {
		certInstalled = "Yes"
	}
	_, certStatus := verifyCert()
	tw.AppendRow(table.Row{"Certificate installed: ", certInstalled})
	tw.AppendRow(table.Row{"Certificate verified: ", certStatus})

	tw.AppendRow(table.Row{"---"})

	envExists := "Exists"
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		envExists = ""
	}

	tw.AppendRows([]table.Row{
		{".env: ", envExists},
		{"$DOCDEV PHPV: ", os.Getenv("PHPV")},
		{"$DOCDEV DOCUMENTROOT: ", os.Getenv("DOCUMENTROOT")},
		{"$DOCDEV TLD_SUFFIX: ", os.Getenv("TLD_SUFFIX")},
	})
	tw.AppendRow(table.Row{"---"})

	hostCmd := "hostctl status docdev -o raw"
	hostOut, _ := exec.Command("bash", "-c", hostCmd).Output()
	hostLines := strings.Split(string(hostOut), "\n")
	if len(hostLines) > 1 {
		hostField := strings.Fields(hostLines[1])
		hostOk := "Error: hosts profile \"docdev\" not enabled."
		if len(hostField) > 1 && hostField[1] == "on" {
			hostOk = "Yes (" + hostField[0] + ")"
		}
		tw.AppendRow(table.Row{"Hosts configured: ", hostOk})
	} else {
		tw.AppendRow(table.Row{"Error: Hosts configured: ", "Profile \"docdev\" not found"})
	}

	mysqlPing, err := ping.NewPinger("mysql")
	mysqlPing.Count = 1
	mysqlPing.OnFinish = func(stats *ping.Statistics) {
		tw.AppendRow(table.Row{"MySQL reachable: ", "Yes (" + stats.IPAddr.String() + ")"})
	}
	err = mysqlPing.Run()
	if err != nil {
		tw.AppendRow(table.Row{"MySQL reachable: ", "Error: " + err.Error()})
	}

	redisPing, err := ping.NewPinger("redis")
	redisPing.Count = 1
	redisPing.OnFinish = func(stats *ping.Statistics) {
		tw.AppendRow(table.Row{"Redis reachable: ", "Yes (" + stats.IPAddr.String() + ")"})
	}
	err = redisPing.Run()
	if err != nil {
		tw.AppendRow(table.Row{"Redis reachable: ", "Error: " + err.Error()})

	}
	tw.AppendRow(table.Row{"---"})

	cmd := "docker-compose ps"
	out, _ := exec.Command("bash", "-c", cmd).Output()
	lines := strings.Split(string(out), "\n")

	for _, line := range deleteEmptySlice(lines) {
		fmtd := strings.Fields(line)
		message := "Running"
		if fmtd[3] != "running" {
			message = "Error: " + fmtd[3]
		}
		if fmtd[2] == "php-fpm" {
			tw.AppendRow(table.Row{"Docker PHP: ", message})
		} else if fmtd[2] == "apache" {
			tw.AppendRow(table.Row{"Docker Apache: ", message})
		} else if fmtd[2] == "bind" {
			tw.AppendRow(table.Row{"Docker Bind: ", message})
		} else if fmtd[2] == "mailhog" {
			tw.AppendRow(table.Row{"Docker Mailhog: ", message})
		}
	}

	tw.SetRowPainter(table.RowPainter(func(row table.Row) txt.Colors {
		if len(row) == 1 {
			return txt.Colors{txt.FgWhite}
		}

		if strings.HasPrefix(fmt.Sprint(row[1]), "Error") {
			return txt.Colors{txt.BgRed, txt.FgBlack}
		}

		switch row[1] {
		case "":
			return txt.Colors{txt.BgRed, txt.FgBlack}
		default:
			return txt.Colors{txt.BgGreen, txt.FgBlack}
		}
	}))

	tw.Render()

	return nil
}

func GenerateCerts(c *cli.Context) error {
	names := getProjectHosts()

	mkCertCmd := "mkcert -cert-file cert/nginx.pem -key-file cert/nginx.key localhost 127.0.0.1 ::1 " + names
	_, err := exec.Command("bash", "-c", mkCertCmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	mkCertPath, err := exec.Command("mkcert", "-CAROOT").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	mkCertPath = mkCertPath[:len(mkCertPath)-1]

	cpCertCmd := `cp -Rf "` + string(mkCertPath[:]) + `"/ ./cert/`
	_, err = exec.Command("bash", "-c", cpCertCmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	fmt.Printf("%s", "Certifcates have been generated.\n")

	if isCertInstalled() == "" {
		fmt.Printf("Root CA is not installed.\n")
		_, err := exec.Command("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", "./cert/rootCA.pem").Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
		fmt.Printf("Root CA has been installed.\n")
	}

	return err
}

func GenerateHosts(c *cli.Context) error {
	hostctl, err := exec.Command("which", "hostctl").Output()
	if string(hostctl[:]) == "" {
		_, err = exec.Command("brew", "install", "guumaster/tap/hostctl").Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
	}

	_, err = exec.Command("hostctl", "backup", "--path", "host/").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		panic(err)
	}

	nameList := getProjectHosts()

	removeHosts := strings.Split(nameList, " ")
	hosts.RemoveHosts(deleteEmptySlice(removeHosts))

	addHosts := strings.ReplaceAll(nameList, ".loc", "."+os.Getenv("TLD_SUFFIX"))
	addHostList := strings.Split(addHosts, " ")
	hosts.AddHosts("127.0.0.1", deleteEmptySlice(addHostList))

	err = hosts.SaveAs("host/modified.hosts")
	if err != nil {
		fmt.Printf("%s", err)
	}

	if c.Bool("dry-run") == false {
		_, err = exec.Command("sudo", "hostctl", "restore", "--from", "host/modified.hosts").Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
	}

	fmt.Printf("%s", "Host file has been generated.\n")

	return err
}

func StartContainer(c *cli.Context) error {

	if c.Bool("php-only") {
		fmt.Printf("Removing existing php-fpm container.\n")
		downCmd := `docker-compose rm -s -f -v php-fpm`
		exec.Command("bash", "-c", downCmd).Output()

		fmt.Printf("Build php-fpm container.\n")
		buildCmd := `docker-compose build php-fpm`
		exec.Command("bash", "-c", buildCmd).Output()
	} else {
		downCmd := `docker-compose rm -s -f -v`
		exec.Command("bash", "-c", downCmd).Output()

		buildCmd := `docker-compose build`
		exec.Command("bash", "-c", buildCmd).Output()
	}

	fmt.Printf("Starting all containers.\n")
	startCmd := `docker-compose up -d`
	start := exec.Command("bash", "-c", startCmd)

	fmt.Printf("%v\n", start)
	start.Stdout = os.Stdout
	start.Stderr = os.Stderr
	start.Stdin = os.Stdin

	err := start.Start()
	if err != nil {
		return err
	}

	if err := start.Wait(); err != nil {
		return err
	}

	if err != nil {
		fmt.Printf("%s", err)
	}

	copyCertsCmd := `docker exec php` + os.Getenv("PHPV") + ` sudo cp -r /etc/ssl/cert/. /etc/ssl/certs/`
	_, err = exec.Command("bash", "-c", copyCertsCmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	refreshCertsCmd := `docker exec php` + os.Getenv("PHPV") + ` sudo update-ca-certificates --fresh`
	_, err = exec.Command("bash", "-c", refreshCertsCmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	if c.Bool("exec") {
		return ExecContainer(c)
	}

	return err
}

func ExecContainer(c *cli.Context) error {
	execCmd := `docker exec -ti php` + os.Getenv("PHPV") + ` zsh`
	cmd := exec.Command("bash", "-c", execCmd)

	env := os.Environ()
	cmd.Env = env

	fmt.Printf("%v\n", cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func ChangePhpVersion(c *cli.Context) error {
	err := setEnvFileValue("PHPV", c.Args().First())

	if c.Bool("start") {
		fmt.Printf("%s", "Starting containers...\n")
		err = StartContainer(c)
		if err != nil {
			return cli.Exit(err, 86)
		}
	}

	// Update the DOCDEV_PHP env for use for other applications
	envVal := "php" + c.Args().First()
	setRcExport("DOCDEV_PHP", envVal)

	return err
}

func setRcExport(variable string, value string) error {
	profileLocation := os.Getenv("HOME") + "/.zshrc"
	if _, err := os.Stat(profileLocation); os.IsNotExist(err) {
		profileLocation = os.Getenv("HOME") + "/.bashrc"
	}

	dat, _ := os.ReadFile(profileLocation)
	split := strings.Split(string(dat), "\n")

	var found bool = false
	for idx, line := range split {
		if strings.HasPrefix(line, "export "+variable) {
			found = true
			re := regexp.MustCompile(`=.*`)
			fix := re.ReplaceAllString(line, "="+value)
			split[idx] = fix
		}
	}

	if !found {
		split = append(split, "export "+variable+"="+value)
	}

	err := ioutil.WriteFile(profileLocation, []byte(strings.Join(split, "\n")), 0)

	os.Setenv(variable, value)

	return err
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func deleteEmptySlice(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func getProjectHosts() string {
	nameCmd := `ls ` + os.Getenv("DOCUMENTROOT") + ` | grep -v / | tr '\n' " " | sed 's/ /\.\l\o\c /g'`
	names, err := exec.Command("bash", "-c", nameCmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	return string(names[:])
}

func isCertInstalled() string {
	certInstalled, err := exec.Command("security", "find-certificate", "-a", "-c", "mkcert").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	return string(certInstalled)
}

func verifyCert() (bool, string) {
	randomHost := strings.Split(getProjectHosts(), " ")

	rootPEM, err := ioutil.ReadFile("./cert/rootCA.pem")
	if err != nil {
		return false, "Error: " + err.Error()
	}

	certPEM, err := ioutil.ReadFile("./cert/nginx.pem")
	if err != nil {
		return false, "Error: " + err.Error()
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		return false, "Error: failed to parse root certificate"
	}

	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return false, "Error: failed to parse certificate PEM"
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, "Error: " + err.Error()
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		DNSName:       randomHost[rand.Intn(len(randomHost))],
		Intermediates: x509.NewCertPool(),
	}

	if _, err := cert.Verify(opts); err != nil {
		return false, "Error: " + err.Error()
	}

	return true, "Verification successful"
}

func setEnvFileValue(key string, value string) error {
	myEnv, err := godotenv.Read()
	if err != nil {
		return err
	}

	myEnv[key] = value
	err = godotenv.Write(myEnv, "./.env")
	if err != nil {
		log.Fatal("Error writing .env file")
	}

	err = godotenv.Overload("./.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return err
}
