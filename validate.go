package main

import (
	"fmt"

	"github.com/getlantern/geneva"
	"github.com/urfave/cli/v2"
)

func init() {
	app.Commands = append(app.Commands,
		&cli.Command{
			Name:  "validate",
			Usage: "validate that a strategy is well-formed",
			Action: func(c *cli.Context) error {
				return validate(c.Args().First())
			},
		})
}

func validate(s string) error {
	strategy, err := geneva.NewStrategy(s)
	if err != nil {
		return cli.Exit(fmt.Sprintf("invalid strategy: %v\n", err), 1)
	}

	fmt.Println(strategy)

	return nil
}
