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
	_, err := db.Exec("INSERT INTO links values(?, ?);", from, to)
	if err != nil {log.Fatal(err)}
}

func getLinkSwitch(cmd, noteID string, db *sql.DB) []note {
	noteList := make([]string, 0)
	switch cmd {
	case "to":
		noteList = getLinksTo(noteID, db)
	case "from":
		noteList = getLinksFrom(noteID, db)
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
