package main

import (
	"fmt"
	"github.com/oklog/ulid/v2"
	"time"
	"flag"
	"strings"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
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
func makeNote(tags string, tagSeperate bool) note {
	mkTime := time.Now().Unix()
	text := createTextEditor("")
	var t string
	if tagSeperate {
		t = createTextEditor(tags)
	} else {
		t = tags
	}
	return note{
							id: ulid.Make().String(),
							created: mkTime,
							modified: mkTime,
							body: string(text),
							tags: validateTags(t)} 
}

// Edit a note
func editNote(n *note, tagSeperate bool) {

	text := createTextEditor(n.body)
	modTime := time.Now().Unix()
	n.modified = modTime
	n.body = text
	var t string
	if tagSeperate {
		t = createTextEditor(strings.Join(n.tags, " "))
		n.tags = validateTags(t)
	}

	return
}

func deleteNoteList(n []note, db *sql.DB) {
	for _, v := range n {
		removeNote(v.id, db)
	}
}


func main() {
	id := flag.String("id", "", "A list of note IDs.")
	tags := flag.String("t", "", "A list of tags.")
	tagSep := flag.Bool("ts", false, "If the user wants to edit tags seperately.")
	dbPath := flag.String("db", "", "Path to the database being used.")
	// TODO add ets (edit-tags-seperately) flag, to, when creating/ eddintg a note, be able to eddint ags in a a seperate text editor instance
	
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 { // quick note
		db := openDB(*dbPath)
		defer db.Close()
		note := makeNote(*tags, *tagSep)
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
			note := makeNote(*tags, *tagSep)
			saveNewNote(&note, db)
			return
		case "search":
			db := openDB(*dbPath)
			defer db.Close()
			n := searchHierarchy(*id, *tags, db)
			printNoteList(n)
		case "edit":
			db := openDB(*dbPath)
			defer db.Close()
			n := searchHierarchy(*id, *tags, db)
			if len(n) != 1 {
				fmt.Println("Sorry, your current options either return 0, of more than 1 note.")
				return
			} else {
				db := openDB(*dbPath)
				defer db.Close()
				editNote(&n[0], *tagSep)
				fmt.Println(&n[0])
				saveNoteUpdate(&n[0], db)
			}
		case "delete":
			db := openDB(*dbPath)
			defer db.Close()
			n := searchHierarchy(*id, *tags, db)
			deleteNoteList(n, db)
		case "tooltip":
			db := openDB(*dbPath)
			defer db.Close()
			editTooltip(*tags, db)	 
		}
	}
}
