package main

import (
	"encoding/json"
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
	// os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)

	fmt.Println(exPath)
	os.Chdir(exPath)
	// wd, _ := os.Getwd()
	// fmt.Println(wd)

	fromFile := func(c *cli.Context) error {
		fmt.Println(app.Name)
		fmt.Println(c.String("command"))

		var savedComs savedCommands
		data, err := os.ReadFile(filepath.Join(exPath, "saved_commands.json"))
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = json.Unmarshal(data, &savedComs)

		if err != nil {
			fmt.Println(err)
		}

		index := -1
		desiredCom := c.String("command")

		for i, s := range savedComs.Items {
			if s.Name == desiredCom {
				index = i
			}
		}

		if index < 0 {
			fmt.Printf("Stored command named %s is not found", desiredCom)
			return nil
		}

		chosenCommand := savedComs.Items[index]

		// fmt.Println(commands)
		realArgs := append([]string{os.Args[0], chosenCommand.CmdStr}, chosenCommand.Args...)
		fmt.Println(realArgs)

		_ = app.Run(realArgs)

		return nil
	}

	app.Commands = append(app.Commands, &cli.Command{
		Name:   "list-adapters",
		Usage:  "Lists the available adapters",
		Action: listAdaptersWrapper,
		Flags:  []cli.Flag{},
	})

	app.Commands = append(app.Commands, &cli.Command{
		Name:   "saved-command",
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
