package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCommands(t *testing.T) {
	one := Command{Name: "One", CmdStr: "intercept", Args: []string{"--hey", "you", "guys"}}
	two := Command{Name: "Two", CmdStr: "intercept", Args: []string{"--wass", "up", "guys"}}

	sc := newSavedCommands()

	sc.add(one)
	sc.add(two)

	assert.Equal(t, 2, len(sc.Items))
}

func TestSaveCommands(t *testing.T) {

	saveFile := filepath.Join("testdata", "test_out.json")

	if _, err := os.Stat(saveFile); errors.Is(err, nil) {
		os.Remove(saveFile)
	}

	one := Command{Name: "One", CmdStr: "intercept", Args: []string{"--hey", "you", "guys"}}
	two := Command{Name: "Two", CmdStr: "intercept", Args: []string{"--wass", "up", "guys"}}

	sc := newSavedCommands()

	sc.add(one)
	sc.add(two)

	err := sc.save(saveFile)
	assert.NoError(t, err)

	_, err = os.Stat(saveFile)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(sc.Items))

	sc_in, err := getSavedCommands(saveFile)
	assert.Nil(t, err)

	assert.Equal(t, sc, sc_in)
	ret, err := sc_in.get("One")

	assert.Nil(t, err)
	assert.Equal(t, one, ret)

	ret2, err2 := sc_in.get("Two")
	assert.Nil(t, err2)
	assert.Equal(t, two, ret2)

}

func TestLoadCommands(t *testing.T) {
	saveFile := filepath.Join("testdata", "test_in.json")

	if _, err := os.Stat(saveFile); errors.Is(err, nil) {
		os.Remove(saveFile)
	}

	one := Command{Name: "One", CmdStr: "intercept", Args: []string{"--hey", "you", "guys"}}
	two := Command{Name: "Two", CmdStr: "intercept", Args: []string{"--wass", "up", "guys"}}

	sc := newSavedCommands()

	sc.add(one)
	sc.add(two)

	sc.save(saveFile)
	assert.Equal(t, 2, len(sc.Items))

	sc_in, err := getSavedCommands(saveFile)
	assert.Nil(t, err)

	assert.Equal(t, sc, sc_in)
	ret, err := sc_in.get("One")

	assert.Nil(t, err)
	assert.Equal(t, one, ret)

	ret2, err2 := sc_in.get("Two")
	assert.Nil(t, err2)
	assert.Equal(t, two, ret2)

}
