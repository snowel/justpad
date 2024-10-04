package main

import (
	"log"
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

	n.tags = pullTags(n.id, db)

	return n
}

func searchByIDs(ids []string, db *sql.DB) []note {
	notes := make([]note, len(ids))

	for i, v := range ids {
		n := searchByID(v, db)
		notes[i] = n
	}

	return notes
}

// Gets a list of unique notes(ids) based on a single tag
func searchByTag(tag string, db *sql.DB) []string {
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

	return list
}

func searchByTags(tags []string, db *sql.DB) []note {
	list := make([]string, 0)
	for _, v := range tags {
		tmp := searchByTag(v, db)
		for _, id := range tmp { // TODO Ugly and probaly slower than it needs to be. Replace arrays with sets (i.e. stringset)
			if !slices.Contains(list, id) {
				list = append(list, id)
			}
		}
	}

	return searchByIDs(list, db)
}

// For a note ID, give the list of tags it is tagged with in the databasse
func pullTags(id string, db *sql.DB) []string {
	taggedRows, err := db.Query("SELECT tag FROM tagged WHERE note = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer taggedRows.Close()

	list := make([]string, 0)

	for taggedRows.Next() {
		
		var tag string
		if err := taggedRows.Scan(&tag); err != nil {
			log.Print(err)
		}
		list = append(list, tag)
	}
	return list
}
