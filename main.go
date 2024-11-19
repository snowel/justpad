package main

import (
	"github.com/oklog/ulid/v2"
	"time"
	"fmt"
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
	id, tags, links, searchMode, sortMode string
	active, pocket bool
	rank, count int
}

func emptyNote() note {
	var n note
	n.id = ulid.Make().String()
	return n
}

// Make note from CMD
func makeNote(tags string, tagSeperate bool, tagEditMode string) note {
	mkTime := time.Now().Unix()
	text := createTextEditor("")
	n := note{
							id: ulid.Make().String(),
							created: mkTime,
							modified: mkTime,
							body: string(text),
							tags: validateTags(tags)} 
	if tagSeperate {
		switchEditTags(&n, tagEditMode)
	}

	return n
}

// Edit a note
func editNote(n *note, tagMode string) {

	text := createTextEditor(n.body)
	modTime := time.Now().Unix()
	n.modified = modTime
	n.body = text
	switchEditTags(n, tagMode)
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
	tagEditMode := flag.String("tem", "normal", "An option for editing tags.")
	tagSep := flag.Bool("ts", false, "If the user wants to edit tags seperately.")
	tagSepNew := flag.Bool("tsn", false, "Shorthand for -ts and -tem newline.")
	tagActive := flag.Bool("ta", false, "If the user wants to overide the t flag value witht he tags of the active note.")
	dbPath := flag.String("db", "", "Path to the database being used.")
	sortMode := flag.String("s", "", "Method by which to sort a list of notes.")
	searchMode := flag.String("m", "hierarchy", "Method by which to sort a list of notes.")
	pocket := flag.Bool("p", false, "Specifiy if the pocket is used for searching.")
	clearPocket := flag.Bool("cp", false, "Clear the pocket before formulating selection and executing command.")
	rank := flag.Int("r", 0, "Specify the rank of the.")
	rankOne := flag.Bool("R", false, "Alias for selecting rank 1. Shorhand for '-r 1'")
	count := flag.Int("c", 0, "Specify the maximum number of notes you want to select.")
	displayFormat := flag.String("f", "", "Format in which notes are printed to output.")

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

	// SHORTHANDS

	// TS + newline mode alias
	if *tagSepNew {
		*tagEditMode = "newline"
		*tagSep = true
	}

	// Rank 1 alias
	if *rankOne {*rank = 1}

	// Rank implies pocket
	if *rank != 0 {*pocket = true}

	//Tags of Active
	if *tagActive {
		n := getActive(db)
		overideTags := searchForTags(n.id, db)
		*tags = strings.Join(overideTags, " ")//TODO Another case we are joinnig the array that will be split into an array... All for simplicities sake.
	}

	// Build the selector struct
	selector := selector{
			id: *id,
			tags: *tags,
			links: *links,
			active: *active,
			searchMode: *searchMode,
			pocket: *pocket,
			count: *count,
			rank: *rank,
			sortMode: *sortMode, }

	// Quick note
	if len(args) == 0 {
		note := makeNote(*tags, *tagSep, *tagEditMode)
		saveNewNote(&note, db)
		pushNoteToPocket(note.id, db)
		return
	}
 
	// Main switch case
	switch args[0] {
	case "new":
		note := makeNote(*tags, *tagSep, *tagEditMode)
		saveNewNote(&note, db)
		pushNoteToPocket(note.id, db)
		return
	case "list", "ls":
		n := searchSwitch(selector, db)
		printNoteList(n, *displayFormat, db)
		pushListToPocket(n, db)
	case "edit", "ed":
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		// TODO fintering single notes could be integrated into the post procssing in search switch, adding a simple bool to the func args
		// But maybe it's cleanre this way
		editNote(&n, *tagEditMode)
		saveNoteUpdate(&n, db)
		pushNoteToPocket(n.id, db)
	case "edit-tags", "edt":
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		switchEditTags(&n, *tagEditMode)
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
	case "clear", "clear-active", "ca":
		clearActive(db)
	case "set-link", "sl": // 2 arg command 
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		setLinkSwitch(args[1], n, db)
		//TODO pocket?
	case "set-link-rank", "slr":
		p := getPocket(db)
		if len(p) < 2 {
			fmt.Println("set-link-rank requires your pocket to have at least 2 notes.")
			return
		}
		setLink(p[0], p[1], db)
	case "set-link-rank-reversed", "slrr":
		p := getPocket(db)
		if len(p) < 2 {
			fmt.Println("set-link-rank-reversed requires your pocket to have at least 2 notes.")
			return
		}
		setLink(p[1], p[0], db)
	case "list-links", "lsl": // 2 arg command
		ns := searchSwitch(selector, db)
		n := filterSingle(ns)
		ns = getLinkSwitch(args[1], n.id, db)
		printNoteList(ns, *displayFormat, db)
		pushListToPocket(ns, db)
	case "merge": // 2 arg command
	// TODO add funciton to check for len of args
		ns := searchSwitch(selector, db)
		mode := ""
		if len(args) >= 2 {mode = args[1]} // arg is optional
		merge(ns, "\n---\njustpad-merge\n---\n", mode, db)//TODO temp hardcoded the sep, eventually, will add optional extra arg
	}
}
