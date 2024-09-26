package main

import (
	"log"
	"strings"

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
