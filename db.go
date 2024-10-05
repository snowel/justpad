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


	TODO - SQL typed date and time for searching and sorting?
	*/

func saveNewNote(note *note, db *sql.DB) {
	// insert
	query := fmt. Sprintf("INSERT INTO notes SELECT '%s', '%s', %d, %d WHERE NOT EXISTS (SELECT 1 FROM notes WHERE id = '%s');", note.id, note.body, note.created, note.modified, note.id)
	// TODO Escaping body text, as an SQL injection could do something, maybe?
	dbQuery(db, query)
	// For each tag in the array insertTagRelation
	insertTagRelations(note.id, note.tags, db)
}

func saveNoteUpdate(note *note, db *sql.DB) {
	query := fmt.Sprintf("UPDATE notes SET body = '%s', modified = %d WHERE id = '%s';", note.body, note.modified, note.id)
	dbQuery(db, query)
	updateTagRelations(note.id, note.tags, db)
}

func insertTagRelation(note, tag string, db *sql.DB) {
	// TODO performance check on returning the tag vs executing the query to create the tag

	// Create tag
	q := fmt.Sprintf("INSERT INTO tags SELECT '%s', '', '' WHERE NOT EXISTS (SELECT 1 FROM tags WHERE tag = '%s');", tag, tag)
	dbQuery(db, q)

	// Inser relation
	q = fmt.Sprintf("INSERT INTO tagged SELECT '%s', '%s' WHERE NOT EXISTS (SELECT 1 FROM tagged WHERE note = '%s' AND tag = '%s');", tag, note, note, tag)
	dbQuery(db, q)
}

func insertTagRelations(note string, tags []string, db *sql.DB) {
	for _, tag := range tags {
		insertTagRelation(note, tag, db)
	}
}

func removeTagRelation(note, tag string, db *sql.DB) {
	_, err := db.Exec("DELETE FROM tagged WHERE tag = '?' AND note = '?';", tag, note)// TODO Not sure this will work!!! Check exec!!
	if err != nil {
			log.Fatal(err)
	}
}

func removeTagRelations(note string, tags []string, db *sql.DB) {
	for _, tag := range tags {
		removeTagRelation(note, tag, db)
	}
}


func updateTagRelations(note string, newTags []string, db *sql.DB) {
	// Search the DB for all the tags currently associated to the note
	oldTags := searchForTags(note, db)

	// Compare the tags twice
	// old tags - new tags =  tags to be removed
	delTags := boolDiff(newTags, oldTags)
	removeTagRelations(note, delTags, db)

	// TODO - discovery question - does it create a measurable performace chance to delte first?
	// new tags - old tags already in the db = tags to be added
	addTags := boolDiff(newTags, oldTags)
	insertTagRelations(note, addTags, db)
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

