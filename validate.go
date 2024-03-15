package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/getlantern/geneva"
	"github.com/urfave/cli/v2"
)

func init() {
	app.Commands = append(app.Commands,
		&cli.Command{
			Name:  "validate",
			Usage: "validate that a strategy is well-formed",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "strategy",
					Aliases: []string{"s"},
					Usage:   "A Geneva `STRATEGY` to validate",
				},
				&cli.StringFlag{
					Name:    "strategyFile",
					Aliases: []string{"f"},
					Usage:   "Validate Geneva strategy from `FILE`",
				},
				&cli.StringFlag{
					Name:    "bulkFile",
					Aliases: []string{"b"},
					Usage:   "Validate many Geneva strategies from `FILE`",
				},
			},
			Action: validateFromCLI,
		})
}

func validateFromCLI(c *cli.Context) error {

	checkFlagsMutualEx := func() bool {

		A := c.String("bulkFile") != ""
		B := c.String("strategy") != ""
		C := c.String("strategyFile") != ""

		return !A && !B && C || A && !B && !C || !A && B && !C
	}

	// command-line options can override some of flashlight's config
	if checkFlagsMutualEx() {
		return cli.Exit("Only one of -strategy, -strategyFile, or -bulkFile should be used", 1)
	}

	s := c.String("strategy")
	if s != "" {
		err := validate(s)

		if err != nil {
			return cli.Exit(err, 1)
		}
		return nil
	}

	s = c.String("strategyFile")
	if s != "" {
		in, err := os.ReadFile(s)
		if err != nil {
			return cli.Exit(fmt.Sprintf("cannot open %s: %v", s, err), 1)
		}
		s = string(in)
		err = validate(s)

		if err != nil {
			return cli.Exit(err, 1)
		}
		return nil

	}

	s = c.String("bulkFile")
	if s != "" {
		bulkValidate(s)
		return nil
	}

	return cli.Exit("No strategy specified, use -strategy, -strategyFile, or -bulkFile", 1)
}

func validate(s string) error {
	strategy, err := geneva.NewStrategy(s)
	if err != nil {
		return cli.Exit(fmt.Sprintf("invalid strategy: %v\n", err), 1)
	}

	fmt.Printf("Valid Strategy:  %v\n", strategy)

	return nil
}

func bulkValidate(strategyFile string) error {

	file, err := os.Open(strategyFile)
	if err != nil {
		return cli.Exit(fmt.Sprintf("cannot open bulk file %s: %v", strategyFile, err), 1)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	total := 0
	failed := 0

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		validateError := validate(text)
		if validateError != nil {
			failed += 1
		}
		total += 1
	}

	return cli.Exit(fmt.Sprintf("%d/%d of known strategies validated (%d failed)", total-failed, total, failed), 1)
}
