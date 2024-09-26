package main

import (
	"fmt"
	"strings"
	"github.com/oklog/ulid/v2"
	"os"
	"os/exec"
	"time"
	"log"
	"flag"
)

type note struct {
	id string

	isTree bool
	child string
	sibling string
	
	created int64
	modified int64
	body string
	tags []string
}

// Make note from CMD
func makeNote(tags string) note {
	mkTime := time.Now().Unix()

	tfile := ulid.Make().String()
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

	return note{
							id: ulid.Make().String(),
							created: mkTime,
							modified: mkTime,
							body: string(text),
							tags: validateTags(tags)} 
}


func main() {
	id := flag.String("id", "", "A list of note IDs.")
	tags := flag.String("t", "", "A list of tags.")
	dbPath := flag.String("db", "", "Path to the database being used.")
	
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 { // quick note
		db := openDB(*dbPath)
		defer db.Close()
		note := makeNote(*tags)
		saveNewNote(&note, db)
		return
	} // quick note


	if len(args) == 1 {

	// TODO Can I open and defer DB here? Technically, but will be funky once we have alt UIs...
		switch args[0] {
		case "init-db":
			initDB(*dbPath)
		case "new":
			db := openDB(*dbPath)
			defer db.Close()
			note := makeNote(*tags)
			saveNewNote(&note, db)
			return
		case "search":
			if *id != "" {
				db := openDB(*dbPath)
				defer db.Close()
				n := searchByIDs(strings.Fields(*id), db)
				fmt.Println(n)
				return
			}
			if *tags != "" {
				db := openDB(*dbPath)
				defer db.Close()
				n := searchByTags(strings.Fields(*tags), db)
				fmt.Println(n)
				return
			}
		case "debug":
			fmt.Println(*dbPath)
			}
	}
}
