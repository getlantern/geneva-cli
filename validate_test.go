package main

import (
	"bufio"
	"os"
	"testing"
)

func TestValidate(t *testing.T) {

	file, err := os.Open("strategies.txt")
	if err != nil {
		t.Errorf("Can't open strategies.txt file")
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	total := 0
	failed := 0

	sout := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)

	for scanner.Scan() {
		text := scanner.Text()
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
