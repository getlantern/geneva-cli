package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"sort"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:                   "geneva",
	Usage:					"Genetic Evasion for windows",
	UseShortOptionHandling: true,
	Commands:               make([]*cli.Command, 0, 4),

}

type Command struct {
	Name   string   `json:"name"`
	CmdStr string   `json:"command"`
	Args   []string `json:"args"`
}

type savedCommands struct {
	Items []Command `json:"saved commands"`
}

func init() {
	fromFile := func(c *cli.Context) error {
		fmt.Println(app.Name)
		fmt.Println(c.String("command"))

		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)

		var savedComs savedCommands
		data, err := ioutil.ReadFile(filepath.Join(exPath,"saved_commands.json"))
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = json.Unmarshal(data, &savedComs)

		if err != nil {
			// print out if error is not nil
			fmt.Println(err)
		}

		index := -1
		desiredCom := c.String("command")

		for i, s := range savedComs.Items {
			fmt.Println(i)
			fmt.Println(s.Name)
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
