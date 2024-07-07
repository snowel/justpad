package main

import (
	"fmt"
	"github.com/oklog/ulid/v2"
	"os"
	"os/exec"
	"time"
	"log"
	"strconv"
	"encoding/csv"
	"strings"
	"flag"
)

type note struct {
	id string

	isTree bool
	child string
	sibling string
	
	create int64
	mod int64
	cont string
	tags []string
}

type conf struct {
	notesFolder string // default ~/.dnote/
	textEditor string // cmd argument for the desired text editor

}

var TagReservedCharacters = []string{";", "!", "^", "@", "*", "{", "}", "=", "|", "\\"}

func removeTag(tags []string, tag string) []string {
	nt := len(tags) - 1
	for i, v := range tags {
		if v == tag {
			tags[i] = tags[nt]
			return tags[:nt]
		}
	}
	return tags
}

// Validate tags takes a string of tags and returns a slice without any invalid tags (i.e. tags will falty character)
func validateTags(tags string) []string {
	bad := []string{}
	all := strings.Fields(tags)
	for _, v := range TagReservedCharacters {
		for _, tag := range all {
			if strings.Contains(tag, v) { 
				bad = append(bad, tag)
				fmt.Println("The tag :: " + tag +" :: is not valie because of the character :: " + v + "::")
			}
		}
	}
	fmt.Println("The following tags will eb removed from the list :", bad)
	for _, v := range bad {
		all = removeTag(all, v)
	}
	return all
}

func makeNote(tags string) note {
	mkTime := time.Now().Unix()

	tfile := ulid.Make().String()
	cmd := exec.Command("nvim", tfile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmdERR := cmd.Run()
	fmt.Println(cmdERR)
	//TODO handle ERR

	text, err := os.ReadFile(tfile)
	if err != nil { log.Fatal(err) }
	err = os.Remove(tfile)
	return note{
							id: ulid.Make().String(),
							create: mkTime,
							mod: mkTime,
							cont: string(text),
							tags: validateTags(tags)} //TODO a little goofy to conver to a string if I'me writing as bytes, but might be best for conceptual consistency
}


// CMD line, add new note to collection
func appendNote(tags string) {
	// make a new note
	note := makeNote(tags)
	// write it to the end of the file
	f, err := os.OpenFile(".dnote", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil { log.Fatal(err) }
	defer f.Close()

	wr := csv.NewWriter(f)
	nt := []string{note.id, strconv.FormatInt(note.create, 16), strconv.FormatInt(note.mod, 16), note.cont, strings.Join(note.tags, " ")}
	err = wr.Write(nt)
	if err != nil { log.Fatal(err) }
	wr.Flush()
	if err = wr.Error(); err != nil { log.Fatal(err) }
}



func main() {
	tags := flag.String("t", "", "A list of tags.")
	
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 { appendNote(*tags) } // quick note

	if len(args) == 1 {
	switch args[0] {
	case "new":
		appendNote(*tags)
	//case "ls":
		//printNotes(sort, count, tags)
		}
	}
}
