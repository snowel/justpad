package main

import (
	"log"
	"strings"
	"slices"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func searchByID(id string, db *sql.DB) note {
	var n note

	if err := db.QueryRow("SELECT * FROM notes WHERE id = ?", id).Scan(&n.id, &n.body, &n.created, &n.modified); err != nil {
		// TODO non-fatal handling for a missing note?
		log.Fatal(err)
	}

	return n
}

func searchByIDs(ids string, db *sql.DB) []note {
	each := strings.Fields(ids)
	notes := make([]note, len(each))

	for i, v := range each {
		n := searchByID(v, db)
		notes[i] = n
	}

	return notes
}

func searchByTag(tag string, db *sql.DB) []note {
	noteRows, err := db.Query("SELECT note FROM tagged WHERE tag = ?", tag)
	if err != nil {
		log.Fatal(err)
	}
	defer noteRows.Close()

	list := make([]string, 0)

	for noteRows.Next() {
		
		var id string
		if err := noteRows.Scan(&id); err != nil {
			log.Print(err)
		}

		if !slices.Contains(list, id) {
			list = append(list, id)
		}
	}

	//TODO This is a copy of searchByIDs, but using searhc by IDs woudl require converting to a string, then back to an array
	// One possible solution is to make the function take an array, then use string.Fileds differtly in the main funtion call.
	notes := make([]note, len(list))

	for i, v := range list {
		n := searchByID(v, db)
		notes[i] = n
	}

	return notes
}
