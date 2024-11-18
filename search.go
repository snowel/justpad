package main

import (
	"log"
	"slices"
	"strings"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Search is about finging notes na dpulling them from the db
// The search funciton in implemented in 3 layers:
// 1. The searhc switch takes a selector and determins the approproate searching mode
// 2. The given search mode function goes throguh each criteria of the selector, either additively or coherently, or sleectively adding notes to export list with the searching functions
// 3. Is the get functions, each producing either a list of notes or a list of noteIDs(strings) based on some searhc criteria

// TODO Make all get funcitons return a list of note IDs, then pull the notes at the end of search switch.
// TODO how about make search swithc encompase the flilter?, return (note, []note) and then have the variou funtions just write only the other?
func searchSwitch(s selector, db *sql.DB) []note {
	noteList := make([]note, 0)
	switch s.searchMode {
	case "hierarchy":
		noteList = searchHierarchy(s.id, s.tags, s.links, s.active, s.pocket, s.rank, db)
	case "optional":
		noteList = searchOptional(s.id, s.tags, s.links, s.active, s.pocket, s.rank, db)
	case "combined":
		noteList = searchCombined(s.id, s.tags, s.links, s.active, s.pocket, s.rank, db)
	default:
		noteList = searchHierarchy(s.id, s.tags, s.links, s.active, s.pocket, s.rank, db)
	}

	// Post-process
	
	// Sort
	if s.sortMode != "" {sortNotesMut(noteList, s.sortMode)}		
	
	// Count
	if s.count != 0 && s.count > 0 {
		if s.count >= len(noteList) {
			log.Print("Count specified is greater than, or equal too, the output. All notes will be passed along.")
		} else {
		noteList = noteList[:s.count]
		}
	}
	return noteList
	
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
			n := make([]note, 1) 
			// TODO this feels extremely bad, knowing that I will later be filtering the slice to just return the single note...
			// A potential "fix" is with the -cs/-C/pre-sorting count flag. If I create a flag that will prevent the searching of too many notes,
			// Commands that take only 1 note can overwrite -C/cs to be 1, then the funcitnoallity of the flag will prevent watsted work... sort of
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


// for commands that require a single note
func filterSingle(ns []note) note {
	if len(ns) != 1 {
		log.Fatal("Sorry, your current options either return 0, of more than 1 note.")// TODO Can these fata logs mess up anything by not closing the db?
	}
	n := ns[0]
	return n
}
