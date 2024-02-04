package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

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

	err = json.Unmarshal(data, savedComs)

	if err != nil {
		fmt.Println(err)
		return new(savedCommands), err
	}
	return savedComs, nil
}

func (sc *savedCommands) getCommand(name string) (Command, error) {

	for _, s := range sc.Items {
		if s.Name == name {
			return s, nil
		}
	}
	return *new(Command), errors.New("No command found")
}
