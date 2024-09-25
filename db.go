package main

import (
	"os"
	"log"
	"fmt"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// TODO is called when the file is not found
func initDB(path string) {
	// TODO help mesage if a db already exists
	if path == "" {
		path, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		path = path + "/.justpad"
		
		os.MkdirAll(path, 0775)
		path = path + "/db"
	}
	
	_, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	
	db := openDB(path)
	defer db.Close()
	
	// Create table for notes
	_, err = db.Exec("CREATE TABLE notes(id text UNIQUE, body text, created int, modified int);")
	if err != nil { log.Fatal( err ) }

	// Create table for directional links
	_, err = db.Exec("CREATE TABLE links(start text, end text);")
	if err != nil { log.Fatal( err ) }

	// Creates table for undirected links
	_, err = db.Exec("CREATE TABLE undirected(note1 text, note2 text);")
	if err != nil { log.Fatal( err ) }

	// Creates table for notes included in the tree
	_, err = db.Exec("CREATE TABLE tree(note text UNIQUE, sybling text, child text);")
	if err != nil { log.Fatal( err ) }

	// Create table for tags
	_, err = db.Exec("CREATE TABLE tags(tag text UNIQUE, tooltip text, functions text);")
	if err != nil { log.Fatal( err ) }

	// Tag relations
	_, err = db.Exec("CREATE TABLE tagged(tag text, note text);")
	if err != nil { log.Fatal( err ) }
}

/*

	TODO - Removal/update of DB can be done 2 ways:
	Upon in memory action, immidiately make equivalent change in DB
	Creat update functions

	Upon update - create a column of flags, set the flag on the opperating note to 0, then which each relation being updated, set the flag to 1, then delet all the relations with the flag still on 0
	Uppon update - a temporary table cna be created, to store the added taggs upons updating 
*/

func saveNewNote(note *note, db *sql.DB) {
	// insert
	query := fmt. Sprintf("INSERT INTO notes SELECT '%s', '%s', %d, %d WHERE NOT EXISTS (SELECT 1 FROM notes WHERE id = '%s');", note.id, note.body, note.created, note.modified, note.id)
	// TODO Escaping body text, as an SQL injection could do something, maybe?
	dbQuery(db, query)
	// For each tag in the array insertTagRelation
	for _, tag := range note.tags {
		insertTagRelation(note.id, tag, db)
	}

}


func insertTagRelation(note, tag string, db *sql.DB) {
	// TODO performace check on returning the tag vs executing the query to create the tag

	// Create tag
	q := fmt.Sprintf("INSERT INTO tags SELECT '%s', '', '' WHERE NOT EXISTS (SELECT 1 FROM tags WHERE tag = '%s');", tag, tag)
	dbQuery(db, q)

	// Inser relation
	q = fmt.Sprintf("INSERT INTO tagged SELECT '%s', '%s' WHERE NOT EXISTS (SELECT 1 FROM tagged WHERE note = '%s' AND tag = '%s');", note, tag, note, tag)
	dbQuery(db, q)
}

// Open and return a database
// REMEBER TO DEFER CLOSE
func openDB(path string) *sql.DB {
	if path == "" {
		p, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		path = p //TODO add the default file strucure to all these defaulting empty path things
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Execute a query on a database
func dbQuery(db *sql.DB, query string) {
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

