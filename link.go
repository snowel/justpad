package main

import (
	"fmt"
	"log"
	"slices"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func setLinkSwitch(cmd string, selectedNote note, db *sql.DB) {
	activeNote := getActive(db)

	if activeNote.id == selectedNote.id {
		fmt.Println("Selected note and active note are the same.")
		return
	}

	switch cmd {
	case "to":
		setLink(selectedNote.id, activeNote.id, db)
	case "from":
		setLink(activeNote.id, selectedNote.id, db)
	}
}
func setLink(from, to string, db *sql.DB) {
	_, err := db.Exec("INSERT INTO links SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM links WHERE start = ? AND end = ?);", from, to, from, to)
	if err != nil {log.Fatal(err)}
}

func getLinkSwitch(cmd, noteID string, db *sql.DB) []note {
	noteList := make([]string, 0)
	switch cmd {
	case "to":
		noteList = getLinksTo(noteID, db)
	case "from":
		noteList = getLinksFrom(noteID, db)
	case "all":
		noteList = getLinksTo(noteID, db)
		n := getLinksFrom(noteID, db)
		stringUnion(&noteList, n)
	default:
		fmt.Println(cmd, " Is not a valid argument for list-links.")
	}
	return searchByIDs(noteList, db)
}

// gets list of IDs of all notes liked to by the selected note
func getLinksFrom(noteID string, db *sql.DB) []string {
	noteRows, err := db.Query("SELECT end FROM links WHERE start = ?", noteID)
	if err != nil { log.Fatal(err) }
	defer noteRows.Close()

	list := make([]string, 0)

	for noteRows.Next() {
		var id string
		if err := noteRows.Scan(&id); err != nil { log.Print(err) }
		if !slices.Contains(list, id) { list = append(list, id) }// TODO refactor this into boolUnion/stringUnion and or use sets
	}
	 return list
}

// gets list of IDs of all notes liking to the selected note
func getLinksTo(noteID string, db *sql.DB) []string {
	noteRows, err := db.Query("SELECT start FROM links WHERE end = ?", noteID)
	if err != nil { log.Fatal(err) }
	defer noteRows.Close()

	list := make([]string, 0)

	for noteRows.Next() {
		var id string
		if err := noteRows.Scan(&id); err != nil { log.Print(err) }
		if !slices.Contains(list, id) { list = append(list, id) }// TODO refactor this into boolUnion/stringUnion and or use sets
	}
	 return list
}

func getLinksFromActive(db *sql.DB) []note {
	activeNote := getActive(db)
	noteList := getLinksFrom(activeNote.id, db) 
	return searchByIDs(noteList, db)
}

func getLinksToActive(db *sql.DB) []note {
	activeNote := getActive(db)
	noteList := getLinksTo(activeNote.id, db) 
	return searchByIDs(noteList, db)
}

// Give receivingNote a copy of all links from givingNote where receivingNote takes the role of givingNote
func inheritLinks(receivingNote, givingNote string, db *sql.DB) {
	// Inherit links to
	inheritLinksTo(receivingNote, givingNote, db)
	// Inherit links from
	inheritLinksFrom(receivingNote, givingNote, db)
} 

// TODO rename to/from links (can't keep track of this jazz!)
// Links from note A -> A's Links
// Links to note A -> A's backlinks
func inheritLinksTo(receivingNote, givingNote string, db *sql.DB) {
	noteRows, err := db.Query("SELECT start FROM links WHERE end = ?", givingNote)
	if err != nil { log.Fatal(err) }
	defer noteRows.Close()

	linksTo := make([]string, 0)

	for noteRows.Next() {
		var id string
		if err := noteRows.Scan(&id); err != nil { log.Print(err) }
		if !slices.Contains(linksTo, id) { linksTo = append(linksTo, id) }
	}
	
	for _, v := range linksTo {
		setLink(v, receivingNote, db)
	}
} 

func inheritLinksFrom(receivingNote, givingNote string, db *sql.DB) {
	noteRows, err := db.Query("SELECT end FROM links WHERE start = ?", givingNote)
	if err != nil { log.Fatal(err) }
	defer noteRows.Close()

	linksFrom := make([]string, 0)

	for noteRows.Next() {
		var id string
		if err := noteRows.Scan(&id); err != nil { log.Print(err) }
		if !slices.Contains(linksFrom, id) { linksFrom = append(linksFrom, id) }
	}
	
	for _, v := range linksFrom {
		setLink(receivingNote, v, db)
	}
} 
