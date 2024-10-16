package main

import (
	"time"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Utills are for commands that do something slightly more specific.
// It's code that probably won't be reused, and often has self-contained functionality (like saving and adding to pocket).



// Merge a list of notes
// TODO newlines not being parsed as the sep argument
// There's a speerator heading the merged note
func merge(mergeList []note, sep, mergeMode string, db *sql.DB) note {
	// If the list is a single note long... not doing anything

	// Otherwise, create a new note
	newNote := emptyNote()// USed to be : var newNote note TODO this would creat issues with the ID!!! CHECK ALL FILES FOR THIS
	newNote.created = mergeList[0].created

	// For each note in the list, combine the text into the new note
	for _, v := range mergeList {
		newNote.body = newNote.body + sep + v.body
		stringUnion(&newNote.tags, v.tags)
		if newNote.created > v.created {newNote.created = v.created}
		switch mergeMode {
		case "all":
			inheritLinks(newNote.id, v.id, db)
		case "links":
			inheritLinksFrom(newNote.id, v.id, db)
		case "backlinks":
			inheritLinksTo(newNote.id, v.id, db)
		default:
			continue
		}
	}

	// Provision the new note
	newNote.modified = time.Now().Unix()
	// Save the note
	saveNewNote(&newNote, db)
	// Add the note to the pocket TODO This is effectively becoming a UI function... perhapse not quite
	pushNoteToPocket(newNote.id, db)
	// Delete the list of notes TODO This seems like a waste of effort, revoign in the first loop seems a lot more efficient, but reusing the functions seesm worth it from a code reading perspective.
	deleteNoteList(mergeList, db)

	return newNote
}
