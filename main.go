package main

import (
	"github.com/oklog/ulid/v2"
	"time"
	"flag"
	"strings"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func t() {
	print("debug")
}

type note struct {
	id string

	isTree bool
	child string
	sibling string
	
	created int64// TODO uint64??
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
	active := flag.Bool("a", false, "Refers to the stored active note, if there is one.")
	tags := flag.String("t", "", "A list of tags.")
	tagSep := flag.Bool("ts", false, "If the user wants to edit tags seperately.")
	dbPath := flag.String("db", "", "Path to the database being used.")
	sortMode := flag.String("s", "", "Method by which to sort a list of notes.")
	pocket := flag.Bool("p", false, "Specifiy if the pocket is used for searching.")
	clearPocket := flag.Bool("cp", false, "Clear the pocket before doing anything else.")
	rank := flag.Int("r", 0, "Specify the rank of the.")
	
	flag.Parse()
	args := flag.Args()


	if *clearPocket {
		db := openDB(*dbPath)
		defer db.Close()
		emptyPocket(db)
	}

	if len(args) == 0 { // quick note
		db := openDB(*dbPath)
		defer db.Close()
		note := makeNote(*tags, *tagSep)
		saveNewNote(&note, db)
		pushNoteToPocket(note.id, db)
		return
	} // quick note

	if len(args) == 1 {
		if args[0] == "init-db" {
			initDB(*dbPath)
		}

		db := openDB(*dbPath)
		defer db.Close()

		// TODO add a split here for expresions returning a single note?... then we'd have to eval multiple times...
		switch args[0] {
		case "new":
			note := makeNote(*tags, *tagSep)
			saveNewNote(&note, db)
			pushNoteToPocket(note.id, db)
			return
		case "list":
			n := searchHierarchy(*id, *tags, *active, *pocket, *rank, db)
			if *sortMode != "" {sortNotesMut(n, *sortMode)}		
			printNoteList(n)
			pushListToPocket(n, db)
		case "edit":
			ns := searchHierarchy(*id, *tags, *active, *pocket, *rank, db)
			n := filterSingle(ns)
			editNote(&n, *tagSep)
			saveNoteUpdate(&n, db)
			pushNoteToPocket(n.id, db)
		case "delete":
			ns := searchHierarchy(*id, *tags, *active, *pocket, *rank, db)
			n := filterSingle(ns)
			removeNote(n.id, db)
		case "delete-all":
			n := searchHierarchy(*id, *tags, *active, *pocket, *rank, db)
			deleteNoteList(n, db)
		case "tooltip":
			editTooltip(*tags, db)	 
		case "set":// requires a single note
			ns := searchHierarchy(*id, *tags, *active, *pocket, *rank, db)
			n := filterSingle(ns)
			setActive(n.id, db)
		case "clear":
			clearActive(db)
		case "debug":
			t()
		}
	}
}
