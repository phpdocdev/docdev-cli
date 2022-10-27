package base

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"utils"

	"github.com/txn2/txeh"
	"github.com/urfave/cli/v2"
)

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
	release := fmt.Sprintf("docdev-%s-%s", runtime.GOOS, runtime.GOARCH)
	if release == "docdev-darwin-arm64" {
		release = "docdev-darwin-amd64"
	}
	
	fmt.Printf("Downloading latest release from github: \x1b[36m%s\x1b[0m\n", release)

	cmd := fmt.Sprintf("gh release download -p \"%s\" --repo \"https://github.com/phpdocdev/docdev\"", release)
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	fmt.Println("Attempting to replace the existing binary")
	os.Rename(release, "docdev")
	newpath, _ := exec.Command("which", "docdev").Output()
	targetPath := strings.Replace(string(newpath), "\n", "", -1)
	if targetPath == "" {
		targetPath = "/usr/local/bin/docdev"
	}

	input, err := ioutil.ReadFile("docdev")
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = ioutil.WriteFile(targetPath, input, 0755)
	if err != nil {
		fmt.Println("Error creating", targetPath)
		fmt.Println(err)
		return err
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
		utils.SetEnvFileValue("TLD_SUFFIX", c.String("tld"))
		utils.SetEnvFileValue("DOCUMENTROOT", c.String("root"))
		utils.SetEnvFileValue("PHPV", c.String("php"))

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

func GenerateCerts(c *cli.Context) error {
	names := utils.GetProjectHosts()

	if _, err := os.Stat(utils.CertPath); os.IsNotExist(err) {
		os.MkdirAll(utils.CertPath, 0755)
	}

	mkCertCmd := "mkcert -cert-file " + utils.CertPath + "/nginx.pem -key-file " + utils.CertPath + "/nginx.key localhost 127.0.0.1 ::1 " + names
	_, err := exec.Command("bash", "-c", mkCertCmd).Output()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	mkCertPath, err := exec.Command("mkcert", "-CAROOT").Output()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	mkCertPath = mkCertPath[:len(mkCertPath)-1]

	cpCertCmd := `cp -Rf "` + string(mkCertPath[:]) + `"/ ./` + utils.CertPath + `/`
	_, err = exec.Command("bash", "-c", cpCertCmd).Output()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	fmt.Printf("%s", "Certifcates have been generated.\n")

	_, err = exec.Command("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", "./"+utils.CertPath+"/rootCA.pem").Output()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("Root CA has been installed.\n")

	return err
}

func GenerateHosts(c *cli.Context) error {

	if _, err := os.Stat(utils.HostPath); os.IsNotExist(err) {
		os.MkdirAll(utils.HostPath, 0755)
	}

	hostctl, _ := exec.Command("which", "hostctl").Output()
	if string(hostctl[:]) == "" {
		_, err := exec.Command("brew", "install", "guumaster/tap/hostctl").Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
	}

	_, err := exec.Command("hostctl", "backup", "--path", utils.HostPath+"/").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		panic(err)
	}

	nameList := utils.GetProjectHosts()

	removeHosts := strings.Split(nameList, " ")
	hosts.RemoveHosts(utils.DeleteEmptySlice(removeHosts))

	addHosts := strings.ReplaceAll(nameList, ".loc", "."+os.Getenv("TLD_SUFFIX"))
	addHostList := strings.Split(addHosts, " ")
	hosts.AddHosts("127.0.0.1", utils.DeleteEmptySlice(addHostList))

	err = hosts.SaveAs(utils.HostPath + "/modified.hosts")
	if err != nil {
		fmt.Printf("%s", err)
	}

	if !c.Bool("dry-run") {
		_, err = exec.Command("sudo", "hostctl", "restore", "--from", utils.HostPath+"/modified.hosts").Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
	}

	fmt.Printf("%s", "Host file has been generated.\n")

	return err
}

func StartContainer(c *cli.Context) error {

	if _, err := os.Stat(utils.CertPath); os.IsNotExist(err) {
		GenerateCerts(c)
	}

	docRoot := os.Getenv("DOCUMENTROOT")
	docRoot = strings.Replace(docRoot, "~", "/Users/"+os.Getenv("USER"), 1)
	if docRoot != "" {
		if _, err := os.Stat(docRoot); os.IsNotExist(err) {
			fmt.Printf("\x1b[31mDOCUMENTROOT path in .env does not exist: \033[0m%s", docRoot)
			return nil
		}
	}

	if c.Bool("php-only") {
		fmt.Printf("Removing existing php-fpm container\n")
		downCmd := `docker-compose rm -s -f -v php-fpm`
		exec.Command("bash", "-c", downCmd).Output()

		fmt.Printf("Build php-fpm container\n")
		buildCmd := `docker-compose build php-fpm`
		exec.Command("bash", "-c", buildCmd).Output()
	} else {
		downCmd := `docker-compose rm -s -f -v`
		exec.Command("bash", "-c", downCmd).Output()

		buildCmd := `docker-compose build`
		exec.Command("bash", "-c", buildCmd).Output()
	}

	fmt.Printf("Starting all containers\n")
	startCmd := `docker-compose up -d`
	start := exec.Command("/bin/bash", "-c", startCmd)

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

	copyCertsCmd := fmt.Sprintf("docker exec php%s sudo cp -r /etc/ssl/cert/. /etc/ssl/certs/", os.Getenv("PHPV"))
	fmt.Println("Copying certificates to the PHP container")
	_, err = exec.Command("bash", "-c", copyCertsCmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	refreshCertsCmd := fmt.Sprintf("docker exec php%s sudo update-ca-certificates --fresh", os.Getenv("PHPV"))
	fmt.Println("Updating the PHP container's certificates")
	_, err = exec.Command("bash", "-c", refreshCertsCmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	if c.Bool("exec") {
		return ExecContainer(c)
	}

	cli.Exit("Started containers.", 0)
	return err
}

func ExecContainer(c *cli.Context) error {
	arg := c.Args().First()
	command := "/bin/bash"
	if arg == "" {
		arg = "php-fpm"
	}

	if arg == "php-fpm" {
		command = "zsh"
	}

	containers, err := utils.GetContainers()

	if err != nil || len(containers) == 0 {
		return cli.Exit(errors.New("docdev containers are not running"), 86)
	}

	for _, container := range containers {
		if arg == container.Labels["com.docker.compose.service"] {
			utils.ExecContainer(container.ID, command)
			return nil
		}
	}

	var allowed []string
	for _, container := range containers {
		allowed = append(allowed, container.Labels["com.docker.compose.service"])
	}

	return cli.Exit(errors.New("\x1b[31mContainer not found. Please choose one of the following:\n\t\x1b[36m"+strings.Join(allowed, ", ")+"\x1b[0m"), 0)
}

func ChangePhpVersion(c *cli.Context) error {
	err := utils.SetEnvFileValue("PHPV", c.Args().First())

	if c.Bool("start") {
		fmt.Printf("%s", "Starting containers...\n")
		err = StartContainer(c)
		if err != nil {
			return cli.Exit(err, 86)
		}
	}

	// Update the DOCDEV_PHP env for use for other applications
	envVal := "php" + c.Args().First()
	utils.SetRcExport("DOCDEV_PHP", envVal)

	return err
}
