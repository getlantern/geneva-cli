package main

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
)

func init() {
	app.Commands = append(app.Commands,
		&cli.Command{
			Name:  "dot",
			Usage: fmt.Sprintf("(unavailable on %s) output the strategy graph as an SVG", runtime.GOOS),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "verbose",
					Aliases: []string{"v"},
				},
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Value:   "output.svg",
				},
			},
			ArgsUsage: "STRATEGY",
			Action:    dot,
		},
	)
}

func dot(c *cli.Context) error {
	return cli.Exit(fmt.Sprintf("The 'dot' subcommand is not supported on %s.", runtime.GOOS), 1)
}
