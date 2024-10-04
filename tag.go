package main

import (
	"strings"
	"fmt"
	"slices"
)

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
