package main

import (
	"github.com/oklog/ulid/v2"
	"time"
	"flag"
	"strings"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)


// Welcome to the dnote/justpad source code!
// At the head of each file, below the package and import statements, I try to include a set of comments describing the logic of the code,
// as well as the architechture that should dictate future code.
//
// dnote's functionality is realitively simple:
// 1. Create and read notes using an external editor
// 2. Add tags, links, tree structure, metadata to those notes
// 3. Manipulated and search those notes based on a list of selection criteria

// The main function works as follows:
// The flags parse the selection values, then they are structured in a struct for cleanliness.
// The main switch/case sets off the approriate set of actions.

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

type selector struct {
	id, tags, links, mode string
	active, pocket bool
	rank int
}

func emptyNote() note {
	var n note
	n.id = ulid.Make().String()
	return n
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
	links := flag.String("lk", "", "Use links from/to the active note as a selector.")
	tags := flag.String("t", "", "A list of tags.")
	tagSep := flag.Bool("ts", false, "If the user wants to edit tags seperately.")
	dbPath := flag.String("db", "", "Path to the database being used.")
	sortMode := flag.String("s", "", "Method by which to sort a list of notes.")
	searchMode := flag.String("m", "hierarchy", "Method by which to sort a list of notes.")
	pocket := flag.Bool("p", false, "Specifiy if the pocket is used for searching.")
	clearPocket := flag.Bool("cp", false, "Clear the pocket before formulating selection and executing command.")
	rank := flag.Int("r", 0, "Specify the rank of the.")
	rankOne := flag.Bool("R", false, "Alias for selecting rank 1. Equivalent to writing: '-r 1'")

	flag.Parse()
	args := flag.Args()

	// Special case: initialize the database
	if len(args) > 0 && args[0] == "init-db" {
		initDB(*dbPath)
		return
	}

	// Open the database
	db := openDB(*dbPath)
	defer db.Close()

	// Selector preprocessor
	// These functions alter the selection flags input based on other flags

	// Clear the pocket
	if *clearPocket {emptyPocket(db)}

	// Rank 1 alias
	if *rankOne {*rank = 1}

	// Rank implies pocket
	if *rank != 0 {*pocket = true}

	// Build the selector struct
	selector := selector{
			id: *id,
			tags: *tags,
			links: *links,
			active: *active,
			mode: *searchMode,
			pocket: *pocket,
			rank: *rank, }

	// Quick note
	if len(args) == 0 {
		note := makeNote(*tags, *tagSep)
		saveNewNote(&note, db)
		pushNoteToPocket(note.id, db)
		return
	}
 
	// Main switch case
	switch args[0] {
	case "new":
		note := makeNote(*tags, *tagSep)
		saveNewNote(&note, db)
		pushNoteToPocket(note.id, db)
		return
	case "list", "ls":
		n := searchSwitch(selector, db)
		if *sortMode != "" {sortNotesMut(n, *sortMode)}		
		printNoteList(n)
		pushListToPocket(n, db)
	case "edit", "ed":
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		editNote(&n, *tagSep)
		saveNoteUpdate(&n, db)
		pushNoteToPocket(n.id, db)
	case "edit-tags", "edt":
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		editTags(&n)
		saveNoteUpdate(&n, db)
		pushNoteToPocket(n.id, db)
	case "delete", "d":
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		removeNote(n.id, db)
	case "delete-list", "dls":
		n := searchSwitch(selector, db)
		deleteNoteList(n, db)
	case "tooltip":
		editTooltip(*tags, db)	 
	case "set", "set-active":
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		setActive(n.id, db)
	case "clear":
		clearActive(db)
	case "set-link", "sl": // 2 arg command 
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		setLinkSwitch(args[1], n, db)
		//TODO pocket?
	case "list-links", "lsl": // 2 arg command
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		ns = getLinkSwitch(args[1], n.id, db)
		printNoteList(ns)
		pushListToPocket(ns, db)
	case "merge": // 2 arg command
	// TODO add funciton to check for len of args
		ns := searchSwitch(selector, db)
		mode := ""
		if len(args) >= 2 {mode = args[1]} // arg is optional
		merge(ns, "\n---\njustpad-merge\n---\n", mode, db)//TODO temp hardcoded the sep, eventually, will add optional extra arg
	}
}
