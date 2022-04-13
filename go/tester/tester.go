package tester

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"utils"

	"github.com/go-ping/ping"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/txn2/txeh"
	"github.com/urfave/cli/v2"
)

func TestConfiguration(c *cli.Context) error {
    tw := table.NewWriter()
    rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
    tw.SetOutputMirror(os.Stdout)

    phpv := os.Getenv("DOCDEV_PHP")
    tw.AppendRows([]table.Row{
        {"$USER DOCDEV_PHP: ", phpv},
        {"$USER DOCDEV_PATH: ", utils.GetRcExport("DOCDEV_PATH")},
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
    if utils.IsCertInstalled() != "" {
        certInstalled = "Yes"
    }
    _, certStatus := verifyCert()
    tw.AppendRow(table.Row{"Certificate installed: ", certInstalled})
    tw.AppendRow(table.Row{"Certificate status: ", certStatus})

    tw.AppendRow(table.Row{"---"})

    envExists := "Exists"
    if _, err := os.Stat(".env"); os.IsNotExist(err) {
        envExists = ""
    }

    projectHosts := strings.Split(utils.GetProjectHosts(), " ")

    tw.AppendRows([]table.Row{
        {".env: ", envExists},
        {"$DOCDEV PHPV: ", os.Getenv("PHPV")},
        {"$DOCDEV DOCUMENTROOT: ", os.Getenv("DOCUMENTROOT")},
        {"$DOCDEV TLD_SUFFIX: ", os.Getenv("TLD_SUFFIX")},
        {"Total Projects: ", len(projectHosts) + 1},
    })
    tw.AppendRow(table.Row{"---"})

    hosts, err := txeh.NewHostsDefault()
    
    if len(projectHosts) > 1 {
        randomHost := string(projectHosts[rand.Intn(len(projectHosts))])
        
        hstFound, hostLine, _ := hosts.HostAddressLookup(randomHost)
        hostOk := "Error: Random project not found in /etc/hosts."
        if hstFound {
            hostOk = "Yes (" + randomHost + " : "+ hostLine +")"
        }
        tw.AppendRow(table.Row{"Hosts configured: ", hostOk})
    } else {
        tw.AppendRow(table.Row{"Hosts configured: ", "Error: No projects for hostnames"})
    }

    mysqlPing, err := ping.NewPinger("mysql")
    mysqlPing.Count = 1
    mysqlPing.OnFinish = func(stats *ping.Statistics) {
        tw.AppendRow(table.Row{"MySQL reachable: ", "Yes (mysql : " + stats.IPAddr.String() + ")"})
    }
    err = mysqlPing.Run()
    if err != nil {
        tw.AppendRow(table.Row{"MySQL reachable: ", "Error: " + err.Error()})
    }

    redisPing, err := ping.NewPinger("redis")
    redisPing.Count = 1
    redisPing.OnFinish = func(stats *ping.Statistics) {
        tw.AppendRow(table.Row{"Redis reachable: ", "Yes (redis : " + stats.IPAddr.String() + ")"})
    }
    err = redisPing.Run()
    if err != nil {
        tw.AppendRow(table.Row{"Redis reachable: ", "Error: " + err.Error()})

    }
    tw.AppendRow(table.Row{"---"})

    cmd := "docker-compose ps"
    out, _ := exec.Command("bash", "-c", cmd).Output()
    lines := strings.Split(string(out), "\n")

    for _, line := range utils.DeleteEmptySlice(lines) {
        fmtd := strings.Fields(line)
        message := "Running"
        if len(fmtd) > 4 && fmtd[3] != "running" {
            message = "Error: " + fmtd[3]
        }

        if len(fmtd) < 3 {
            continue
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

    tw.SetRowPainter(table.RowPainter(func(row table.Row) text.Colors {
        if len(row) == 1 {
            return text.Colors{text.FgWhite}
        }

        if strings.HasPrefix(fmt.Sprint(row[1]), "Error") {
            return text.Colors{text.BgRed, text.FgBlack}
        }

        switch row[1] {
        case "":
            return text.Colors{text.BgRed, text.FgBlack}
        default:
            return text.Colors{text.BgGreen, text.FgBlack}
        }
    }))

    tw.Render()

    return nil
}


func verifyCert() (bool, string) {
	randomHost := strings.Split(utils.GetProjectHosts(), " ")

	rootPEM, err := ioutil.ReadFile("./"+utils.CertPath+"/rootCA.pem")
	if err != nil {
		return false, "Error: " + err.Error()
	}

	certPEM, err := ioutil.ReadFile("./"+utils.CertPath+"/nginx.pem")
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
