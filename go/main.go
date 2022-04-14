package main

import (
	"base"
	"fmt"
	"log"
	"os"
	"tester"
	"utils"

	"github.com/urfave/cli/v2"
)

var Version = "v0.0-dev"

func main() {
	if !utils.Setup() {
		return
	}

	flags := []cli.Flag{}

	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Version}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   ` + "\x1b[32m" + `{{join .Names ", "}}` + "\x1b[0m" + `{{"\t"}}` + "\x1b[94m" + `{{.Usage}}` + "\x1b[0m" + `{{"\n"}}{{if .Flags}}{{range .Flags}}` + "\x1b[2m" + `{{"     "}}--{{join .Names ", "}}{{"\t"}}{{.GetUsage}} {{if .GetValue}}(default: {{.GetValue}}){{end}}` + "\x1b[0m" + `{{ "\n"}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	app := &cli.App{
		Name:                 "docdev",
		Version:              Version,
		EnableBashCompletion: true,

		Flags: flags,
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize configuration and install mkcert",
				Action:  base.Init,
				Flags: append([]cli.Flag{
					&cli.StringFlag{
						Name:  "tld",
						Value: "loc",
						Usage: "TLD for project hostnames",
					},
					&cli.StringFlag{
						Name:  "root",
						Value: os.Getenv("HOME") + "/repos/",
						Usage: "Root directory containing your projects",
					},
					&cli.StringFlag{
						Name:  "php",
						Value: "74",
						Usage: "Initial PHP version",
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
				Action:  base.GenerateCerts,
			},
			{
				Name:    "hosts",
				Aliases: []string{},
				Usage:   "Generate a new hosts profile and add it to your system /etc/host",
				Action:  base.GenerateHosts,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Dry run",
					},
				},
			},
			{
				Name:    "start",
				Aliases: []string{"s"},
				Usage:   "Bring up the docker containers",
				Action:  base.StartContainer,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "php-only",
						Usage: "Reset the PHP container",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "exec",
						Usage: "Start container shell after starting",
					},
				},
			},
			{
				Name:    "exec",
				Aliases: []string{"e"},
				Usage:   "Start docker container shell",
				Action:  base.ExecContainer,
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Test your configuration.",
				Action:  tester.TestConfiguration,
			},
			{
				Name:    "php",
				Aliases: []string{"p"},
				Usage:   "Change php version (requires \"start\" to rebuild). Valid values: 54, 56, 72, 74",
				Action:  base.ChangePhpVersion,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "start",
						Usage: "Start the containers after switching the PHP version",
					},
					&cli.BoolFlag{
						Name:  "php-only",
						Usage: "Reset the PHP container",
						Value: true,
					},
				},
			},
			{
				Name:   "refresh",
				Usage:  "Pull changes from git and images from Docker",
				Action: base.Refresh,
			},
			{
				Name:   "selfupdate",
				Usage:  "Update docdev binary, requires \"gh\" to be installed",
				Action: base.SelfUpdate,
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
