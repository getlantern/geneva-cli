package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

var (
	lastExitCode = 0
	fakeOsExiter = func(rc int) {
		lastExitCode = rc
	}
	fakeErrWriter = &bytes.Buffer{}
)

func TestValidate(t *testing.T) {

	fi := filepath.Join("testdata", "strategies.txt")
	file, err := os.Open(fi)
	assert.Nil(t, err, "Can't open strategies.txt file")

	defer file.Close()

	scanner := bufio.NewScanner(file)
	total := 0
	failed := 0

	sout := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)

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
	os.Stdout = sout

	if failed > 0 {
		t.Errorf("%d/%d of known strategies did not validate", failed, total)
	}

}

func TestValidateCLIStrategyStringValid(t *testing.T) {

	exitCode := 0
	called := false

	cli.OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { cli.OsExiter = fakeOsExiter }()

	args := []string{"geneva-cli", "validate", "-s", "[TCP:flags:PA]-fragment{tcp:-1:True}-| \\/"}

	_ = app.Run(args)

	assert.Equal(t, 0, exitCode)
	assert.False(t, called)
}

func TestValidateCLIStrategyStringInvalid(t *testing.T) {

	exitCode := 0
	called := false

	cli.OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { cli.OsExiter = fakeOsExiter }()

	sout := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)

	args := []string{"geneva-cli", "validate", "-s", "[TCP:flags:PA]-fragment{sdgdfghdfghjerdastgtcp:-1:True}-| \\/"}

	os.Stdout = sout

	_ = app.Run(args)

	assert.Equal(t, 1, exitCode)
	assert.True(t, called)

}

func TestValidateCLIStrategyFileValid(t *testing.T) {

	exitCode := 0
	called := false

	cli.OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { cli.OsExiter = fakeOsExiter }()
	sout := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)

	fi := filepath.Join("testdata", "s.txt")

	args := []string{"geneva-cli", "validate", "-strategyFile", fi}

	os.Stdout = sout

	_ = app.Run(args)

	assert.Equal(t, 0, exitCode)
	assert.False(t, called)

}

func TestValidateCLIStrategyBulkValid(t *testing.T) {

	exitCode := 0
	called := false

	cli.OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { cli.OsExiter = fakeOsExiter }()

	sout := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	fi := filepath.Join("testdata", "strategies.txt")
	args := []string{"geneva-cli", "validate", "-b", fi}

	_ = app.Run(args)

	os.Stdout = sout

	assert.Equal(t, 0, exitCode)
	assert.False(t, called)
}
