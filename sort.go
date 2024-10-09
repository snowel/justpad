package main

import (
	"sort"
)

// Sorts a slice of notes. Muatative.
func sortNotesMut(n []note, mode string) {
	switch mode{
	case "modified-new":
		sort.Slice(n, func(i, j int) bool {return n[i].modified > n[j].modified})
	case "modified-old":
		sort.Slice(n, func(i, j int) bool {return n[i].modified < n[j].modified})
	case "created-new":
		sort.Slice(n, func(i, j int) bool {return n[i].created > n[j].created})
	case "created-old":
		sort.Slice(n, func(i, j int) bool {return n[i].created < n[j].created})
	case "size-s":
		sort.Slice(n, func(i, j int) bool {return len(n[i].body) < len(n[j].body)})
	case "size-l":
		sort.Slice(n, func(i, j int) bool {return len(n[i].body) > len(n[j].body)})
	default:
		return
	}
}
