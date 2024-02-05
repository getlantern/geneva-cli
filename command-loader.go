package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func init() {

	app.Commands = append(app.Commands, &cli.Command{
		Name:   "list-saved-commands",
		Usage:  "Lists the saved-commands",
		Action: listSavedCommands,
		Flags:  []cli.Flag{},
	})
}

type Command struct {
	Name   string   `json:"name"`
	CmdStr string   `json:"command"`
	Args   []string `json:"args"`
}

type savedCommands struct {
	Items []Command `json:"saved commands"`
}

func getSavedCommands(commandFilepath string) (*savedCommands, error) {
	var savedComs *savedCommands
	data, err := os.ReadFile(commandFilepath)
	if err != nil {
		fmt.Println(err)
		return new(savedCommands), err
	}

	err = json.Unmarshal(data, &savedComs)

	if err != nil {
		fmt.Println(err)
		return new(savedCommands), err
	}
	return savedComs, nil
}

func (sc *savedCommands) get(name string) (Command, error) {

	for _, s := range sc.Items {
		if s.Name == name {
			return s, nil
		}
	}
	return *new(Command), errors.New("no command found")
}

func listSavedCommands(c *cli.Context) error {

	sc, err := getSavedCommands("saved_commands.json")
	if err != nil {
		return cli.Exit(fmt.Sprintf("Could not read command file: %v\n", err), 1)
	}

	for _, v := range sc.Items {
		fmt.Printf("Name: %s\nCommand: %s %v \n", v.Name, v.CmdStr, v.Args)
	}
	return nil
}

func (sc *savedCommands) add(newCom Command) {
	sc.Items = append(sc.Items, newCom)
}

func (sc *savedCommands) save(filePath string) error {
	jsonString, _ := json.Marshal(sc)
	return os.WriteFile(filePath, jsonString, os.ModePerm)
}
