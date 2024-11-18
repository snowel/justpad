package main

import (
	"fmt"
	"time"
	"strings"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

//TODO Make it pretty?

// Switch switches the formating function for printing a SINGLE note
func printSwitch(n *note, mode string, db *sql.DB) {
	switch mode {
		case "simple", "s":
			printNoteSimple(n)
		case "links-in":// Prints indetn notes which are linked to the given note and included within the search
		case "links", "l":// Prints all
			printNoteWithLinks(n, db)
		case "for-links":// INTERNAL USE -- Used by the formating function to print the indented list --
			printNoteWithIndent(n, "|----------", "", "|       ")
		default:
			printNote(n)
	}
}

// --- Individual styles to format a SINGLE note ---
func printNote(n *note) {
	fmt.Println(n.id)
	fmt.Println("Created: ", time.Unix(n.created, 0).String(), "Modified: ", time.Unix(n.modified, 0).String())
	fmt.Println(n.tags, "\n")
	fmt.Println(n.body)
}

func printNoteWithIndent(n *note, head, foot, indent string) {
	fmt.Println(head, n.id)
	fmt.Println(indent, "Created: ", time.Unix(n.created, 0).String(), "Modified: ", time.Unix(n.modified, 0).String())
	fmt.Println(indent, n.tags)//TODO Lots of tags wont indent properly
	fmt.Println(indent)//TODO Lots of tags wont indent properly
	fmt.Println(indent, strings.ReplaceAll(n.body, "\n", "\n" + indent))
	if foot != "" {fmt.Println(foot)}
}

func printNoteSimple(n *note) {
	fmt.Println(n.body)
}


func printNoteWithLinks(n *note, db *sql.DB) {
	printNote(n)
	ns := getLinkSwitch("from", n.id, db)
	printNoteListClean(ns, "for-links", db)
	fmt.Println("")
}


// --- Print multiple notes

func printNoteList(ns []note, mode string, db *sql.DB) {
	for i, v := range ns {
		fmt.Println("///////// Rank: ", i+1)
		printSwitch(&v, mode, db)
	}
}

// Print a list of notes without rank or delimiter
func printNoteListClean(ns []note, mode string, db *sql.DB) {
	for _, v := range ns {
		printSwitch(&v, mode, db)
	}
}
