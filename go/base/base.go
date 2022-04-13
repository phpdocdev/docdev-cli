package base

import (
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

	if utils.IsCertInstalled() == "" {
		fmt.Printf("Root CA is not installed.\n")
		_, err := exec.Command("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", "./"+utils.CertPath+"/rootCA.pem").Output()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		fmt.Printf("Root CA has been installed.\n")
	}

	return err
}

func GenerateHosts(c *cli.Context) error {

	if _, err := os.Stat(utils.HostPath); os.IsNotExist(err) {
		os.MkdirAll(utils.HostPath, 0755)
	}

	hostctl, err := exec.Command("which", "hostctl").Output()
	if string(hostctl[:]) == "" {
		_, err = exec.Command("brew", "install", "guumaster/tap/hostctl").Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
	}

	_, err = exec.Command("hostctl", "backup", "--path", utils.HostPath+"/").Output()
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

	if c.Bool("dry-run") == false {
		_, err = exec.Command("sudo", "hostctl", "restore", "--from", utils.HostPath+"/modified.hosts").Output()
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
