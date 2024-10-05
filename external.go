package main

import (
	"os/exec"
	"os"
	"log"
	"github.com/oklog/ulid/v2"
	
)

func createTextEditor(s string) string {
	tfile := ulid.Make().String()
	f, err := os.Create(tfile)// TODO review FileMode use
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(s)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("nvim", tfile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmdERR := cmd.Run()
	if cmdERR != nil {
		log.Fatal(cmdERR)
	}

	text, err := os.ReadFile(tfile)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(tfile)
	if err != nil {
		log.Fatal(err)
	}
	return string(text)
}

// Code calling other binaries/commans to edit/view files
