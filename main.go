package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:                   "geneva",
	UseShortOptionHandling: true,
	Commands:               make([]*cli.Command, 0, 4),
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
