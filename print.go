package main

import (
	"fmt"
	"time"
)

//TODO Make it pretty, i.e. bubbletea
func printNote(n *note) {
	fmt.Println(n.id)
	fmt.Println("Created: ", time.Unix(n.created, 0).String(), "Modified: ", time.Unix(n.modified, 0).String())
	fmt.Println(":: note ::")
	fmt.Println(n.body)
	fmt.Println(":: end ::")
	fmt.Println("Tagged with:", n.tags)
}

func printNoteList(ns []note) {
	for i, v := range ns {
		fmt.Println("Rank: ", i+1)
		printNote(&v)
	}
}
