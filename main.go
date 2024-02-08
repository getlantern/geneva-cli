package main

import (
	"fmt"
	"os"
	"path/filepath"

	"sort"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:                   "geneva",
	Usage:                  "Genetic Evasion for windows",
	UseShortOptionHandling: true,
	Commands:               make([]*cli.Command, 0, 4),
}

func init() {
	exPath := getExecPath()
	os.Chdir(exPath)

	fromFile := func(c *cli.Context) error {
		fmt.Println(app.Name)
		fmt.Println(c.String("command"))

		sc, err := getSavedCommands("saved_commands.json")
		if err != nil {
			return cli.Exit(fmt.Sprintf("Could not read command file: %v\n", err), 1)
		}

		desiredCom := c.String("command")

		chosenCommand, err := sc.get(desiredCom)
		if err != nil {
			fmt.Printf("Stored command named %s is not found", desiredCom)
			return nil
		}

		realArgs := append([]string{os.Args[0], chosenCommand.CmdStr}, chosenCommand.Args...)
		fmt.Println(realArgs)

		_ = app.Run(realArgs)

		return nil
	}

	app.Commands = append(app.Commands, &cli.Command{
		Name:   "load-command",
		Usage:  "Runs commands from config file",
		Action: fromFile,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "command",
				Aliases:  []string{"c"},
				Value:    "default",
				Required: true,
			},
		},
	})

}

func main() {
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	for _, c := range app.Commands {
		sort.Sort(cli.FlagsByName(c.Flags))
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func getExecPath() string {
	ex, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	exPath := filepath.Dir(ex)
	return exPath
}
