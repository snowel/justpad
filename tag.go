package main

import (
	"strings"
	"fmt"
	"slices"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type tag struct {
	tag string
	tooltip string
	functions string
}

var TagReservedCharacters = []string{";", "!", "^", "@", "*", "{", "}", "=", "|", "\\"}

func removeTag(tags []string, tag string) []string {
	nt := len(tags) - 1
	for i, v := range tags {
		if v == tag {
			tags[i] = tags[nt]
			return tags[:nt]
		}
	}
	return tags
}

// Validate tags takes a string of tags and returns a slice without any invalid tags 
func validateTags(tags string) []string {
	bad := []string{}
	all := strings.Fields(tags)
	for _, v := range TagReservedCharacters {
		for _, tag := range all {
			if strings.Contains(tag, v) { 
				bad = append(bad, tag)
				fmt.Println("The tag :: " + tag +" :: is not valid because of the character :: " + v + "::")
			}
		}
	}
	if len(bad) > 0 {
		fmt.Println("The following tags will be removed from the list : ", bad)
		for _, v := range bad {
			all = removeTag(all, v)
		}
	}

	return all
}

// --- functions for combining UI funcitons related to tags
 
// TODO refactor getting a single valid tag to a filter function
func editTooltip(tag string, db *sql.DB) {
	if tag == "" {
		fmt.Println("You must provide a single, valid tag to edit the tooltip.")
		return
	}
	ts := validateTags(tag)
	nTags := len(ts)
	if nTags > 1 {
		fmt.Println("Please provide only 1 tag.")
		return
	}
	if nTags < 1 {
		fmt.Println("Please provide 1 valid tag.")
		return
	}

	t := getTag(ts[0], db)

	if t.tag == "" {
		fmt.Println("This tag doesn't exist.")
		return
	}
	newTooltip := createTextEditor(t.tooltip)

	t.tooltip = newTooltip
	saveTagUpdate(&t, db)

}

// --- Operations on slices of tags

// Boolean difference
func boolDiff(first, second []string) []string {
	res := make([]string, 0)
	for _, v := range first {//TODO There is certainly a cuter, and faster, way to do this
		if !slices.Contains(second, v) {
			res = append(res, v)
		}
	}
	return res
} 

// Adds addSlice to mutSlice, without creating duplicates 
// TODO Ugly and probaly slower than it needs to be. Replace arrays with sets (i.e. stringset)
func stringUnion(mutSlice *[]string, addSlice []string) {//TODO testing the ergonomics of having pointer to only the mut slice 
	for _, v := range addSlice { 
		if !slices.Contains(*mutSlice, v) {
			*mutSlice = append(*mutSlice, v)
		}
	}
}

// TODO generics or generalization
// Adds addSlice to mutSlice, without creating duplicates 
func noteUnion(mutSlice, addSlice *[]note) {
	for _, v :=  range *addSlice {
		if !slices.ContainsFunc(*mutSlice, func(n note) bool {return n.id == v.id}) {
			*mutSlice = append(*mutSlice, v)
		}
	}
}

// returns a slice of all notes in common between the 2 notes
func noteIntersect(as, bs []note) []note {
	cs := make([]note, 0)
	for _, v :=  range bs {
		if slices.ContainsFunc(as, func(n note) bool {return n.id == v.id}) {
			cs = append(cs, v)
		}
	}
	return cs
}
