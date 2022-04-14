package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/joho/godotenv"
)

const (
	CertPath = "data/cert"
	HostPath = "data/hosts"
)

func Setup() bool {
	docDevPath := GetRcExport("DOCDEV_PATH")
	if docDevPath == "" {
		fmt.Printf("%s", text.Color(text.FgRed).Sprint("Error: missing DOCDEV_PATH from your current environment.\n\n"))
		fmt.Printf("%s", text.Color(text.FgYellow).Sprint("echo \"export DOCDEV_PATH=/Users/$USER/docdev\" >> "+GetProfileLocation()+"\n"))
		return false
	}

	if _, err := os.Stat(docDevPath); os.IsNotExist(err) {
		os.MkdirAll(docDevPath, 0755)
	}

	os.Chdir(docDevPath)

	if IsDirEmpty(docDevPath) {
		fmt.Println("\x1b[94mCloning docker-dev to " + docDevPath + "...\033[0m")
		cmd := "git clone https://github.ark.org/brandon-kiefer/docker-dev.git ."
		_, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			fmt.Println(err)
		}

		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			fmt.Println("\x1b[94mCopying .env.example to .env\033[0m")
			_, err := exec.Command("cp", ".env.example", ".env").Output()
			if err != nil {
				fmt.Printf("%s", err)
			}
		}
	}

	LoadEnv()
	return true
}

func GetProfileLocation() string {
	defaultLocation := os.Getenv("HOME") + "/.zshrc"
	if _, err := os.Stat(defaultLocation); os.IsNotExist(err) {
		return os.Getenv("HOME") + "/.bashrc"
	}
	return defaultLocation
}

func GetRcExport(variable string) string {
	envVal := os.Getenv(variable)
	if envVal == "" {
		profileLocation := GetProfileLocation()
		dat, _ := os.ReadFile(profileLocation)
		split := strings.Split(string(dat), "\n")
		for _, line := range split {
			if strings.HasPrefix(line, "export "+variable) {
				fix := strings.Split(line, "=")
				return fix[1]
			}
		}
	}

	return envVal
}

func SetRcExport(variable string, value string) error {
	profileLocation := GetProfileLocation()

	dat, _ := os.ReadFile(profileLocation)
	split := strings.Split(string(dat), "\n")

	var found string = GetRcExport(variable)

	if found == "" {
		split = append(split, "export "+variable+"="+value)
	}

	err := ioutil.WriteFile(profileLocation, []byte(strings.Join(split, "\n")), 0)

	os.Setenv(variable, value)

	return err
}

func SetEnvFileValue(key string, value string) error {
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

func IsCertInstalled() string {
	certInstalled, err := exec.Command("security", "find-certificate", "-a", "-c", "mkcert").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	return string(certInstalled)
}

func GetProjectHosts() string {
	nameCmd := `ls ` + os.Getenv("DOCUMENTROOT") + ` | grep -v / | tr '\n' " " | sed 's/ /\.\l\o\c /g'`
	names, err := exec.Command("bash", "-c", nameCmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	return string(names[:])
}

func DeleteEmptySlice(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func IsDirEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true
	}
	return false // Either not empty or error, suits both cases
}
