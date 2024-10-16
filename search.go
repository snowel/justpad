package main

import (
	"log"
	"slices"
	"strings"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func searchSwitch(s selector, db *sql.DB) []note {
	switch s.mode {
	case "hierarchy":
		return searchHierarchy(s.id, s.tags, s.links, s.active, s.pocket, s.rank, db)
	case "optional":
		return searchOptional(s.id, s.tags, s.links, s.active, s.pocket, s.rank, db)
	case "combined":
		return searchCombined(s.id, s.tags, s.links, s.active, s.pocket, s.rank, db)
	default:
		return searchHierarchy(s.id, s.tags, s.links, s.active, s.pocket, s.rank, db)
	}
	
}

func searchCombined(id, tags, links string, active, pocket bool, rank int, db *sql.DB) []note {
	ns := make([]note, 0)
	firstNoteFlag := true// If we don't chekc for a first note, our empty slice will cancel every note we find
	if active {
		singleNote := getActive(db)
		n := make([]note, 1) // TODO this feels extremely bad, knowing that I will later be filtering the slice to just return the single note...
		n[0] = singleNote
		if firstNoteFlag {
			ns = n
			firstNoteFlag = false
		}// There is no posibility of it not being the first note here
	}

	if id != "" {
		n := searchByIDs(strings.Fields(id), db)
		if firstNoteFlag {
			ns = n
			firstNoteFlag = false
		} else {
			ns = noteIntersect(ns, n)
		}
	}
	if links != "" {
		switch links {
		case "to":
			n := getLinksToActive(db)
			if firstNoteFlag {
				ns = n
				firstNoteFlag = false
			} else {
				ns = noteIntersect(ns, n)
			}
			noteUnion(&ns, &n)
		case "from":
			n := getLinksFromActive(db)
			if firstNoteFlag {
				ns = n
				firstNoteFlag = false
			} else {
				ns = noteIntersect(ns, n)
			}
		}
	}
	if pocket {
		if rank != 0 {
			singleNote := searchSinglePocket(rank, db)
			n := make([]note, 1) // TODO this feels extremely bad, knowing that I will later be filtering the slice to just return the single note...
			n[0] = singleNote
			if firstNoteFlag {
				ns = n
				firstNoteFlag = false
			} else {
				ns = noteIntersect(ns, n)
			}
		} else {
			n := searchFullPocket(db)
			if firstNoteFlag {
				ns = n
				firstNoteFlag = false
			} else {
				ns = noteIntersect(ns, n)
			}
		}
	}
	if tags != "" {
		n := searchByTags(strings.Fields(tags), db)
		if firstNoteFlag {
			ns = n
			firstNoteFlag = false
		} else {
			ns = noteIntersect(ns, n)
		}
	}
	return ns
}

// TODO this can probably be quicker
func searchOptional(id, tags, links  string, active, pocket bool, rank int, db *sql.DB) []note {
	ns := make([]note, 0)
	if active {
		n := getActive(db)
		notes := make([]note, 1) // TODO this feels extremely bad, knowing that I will later be filtering the slice to just return the single note...
		notes[0] = n
		noteUnion(&ns, &notes)
	}

	if id != "" {
		n := searchByIDs(strings.Fields(id), db)
		noteUnion(&ns, &n)
	}
	if links != "" {
		switch links {
		case "to":
			n := getLinksToActive(db)
			noteUnion(&ns, &n)
		case "from":
			n := getLinksFromActive(db)
			noteUnion(&ns, &n)
		}
	}
	if pocket {
		if rank != 0 {
			n := searchSinglePocket(rank, db)
			notes := make([]note, 1) // TODO this feels extremely bad, knowing that I will later be filtering the slice to just return the single note...
			notes[0] = n
			noteUnion(&ns, &notes)
		} else {
			n := searchFullPocket(db)
			noteUnion(&ns, &n) 
		}
	}
	if tags != "" {
		n := searchByTags(strings.Fields(tags), db)
		noteUnion(&ns, &n)
	}
	return ns
}

// Takes possible search methods and executes them in order of priority TODO Add combined search (althgouh that might require experaiion, or, proprecisely, another flag (i.e. all matching or matching all (i.e. tags and creation date, or tags/or creation date)))
// Currently, this is pure hierarchy ID, then tags
func searchHierarchy(id, tags, links string, active, pocket bool, rank int, db *sql.DB) []note {
	if active {
		n := getActive(db)
	notes := make([]note, 1) // TODO this feels extremely bad, knowing that I will later be filtering the slice to just return the single note...
	notes[0] = n
	return notes
	}
	if id != "" {
		n := searchByIDs(strings.Fields(id), db)
		return n
	}
	if links != "" {
		switch links {
		case "to":
			return getLinksToActive(db)
		case "from":
			return getLinksFromActive(db)
		}
	}
	if pocket {
		if rank != 0 {
			n := searchSinglePocket(rank, db)
			notes := make([]note, 1) // TODO this feels extremely bad, knowing that I will later be filtering the slice to just return the single note...
			notes[0] = n
			return notes
		}
		return searchFullPocket(db)
	}
	if tags != "" {
		n := searchByTags(strings.Fields(tags), db)
		return n
	}
	return make([]note, 0)
}

func searchByID(id string, db *sql.DB) note {
	var n note

	if err := db.QueryRow("SELECT * FROM notes WHERE id = ?", id).Scan(&n.id, &n.body, &n.created, &n.modified); err != nil {
		// TODO non-fatal handling for a missing note?
		log.Fatal(err)
	}

	n.tags = searchForTags(n.id, db)

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
		stringUnion(&list, searchByTag(v, db))
	}
	return searchByIDs(list, db)
}

// For a note ID, give the list of tags it is tagged with in the databasse
func searchForTags(id string, db *sql.DB) []string {
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

// for commands that require a single note
func filterSingle(ns []note) note {
	if len(ns) != 1 {
		log.Fatal("Sorry, your current options either return 0, of more than 1 note.")// TODO Can these fata logs mess up anything by not closing the db?
	}
	n := ns[0]
	return n
}
