package main

import (
	"log"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Active note
// Active note is a single note that can be manually set asside for later use/reference

func getActive(db *sql.DB) note {
	var noteID string
	if err := db.QueryRow("SELECT note FROM active;").Scan(&noteID); err != nil { 
		log.Print("There is no active note.")
		log.Fatal(err)
	}
	return searchByID(noteID, db)	
}

func setActive(noteID string, db *sql.DB) {
	clearActive(db)
	_, err := db.Exec("INSERT INTO active values(?);", noteID)
	if err != nil { log.Fatal( err ) }
}

func clearActive(db *sql.DB) {
	_, err := db.Exec("DELETE FROM active;")
	if err != nil { log.Fatal( err ) }
}

// Live pocket - live pocket is dedicated to memory only, used during a read mode session

// Persistent pocket - persistent pocket is stored in db, used through CLI
// TODO maybe pocket should be distinct from memory. The current impl of pocket is memory, where anything you intract with is added to the pocket, where as deliberately addign things to store for later use would be a seperate thing

func pushNoteToPocket(noteID string, db *sql.DB) {
	var rank int
	if err := db.QueryRow("SELECT rank FROM pocket WHERE note = ?;", noteID).Scan(&rank); err != nil {
		if err == sql.ErrNoRows {
			// note is not pocketed -> increment everything's rank by 1, add note at rank 1
			_, err = db.Exec("UPDATE pocket SET rank = rank + 1;")
			if err != nil { log.Fatal( err ) }
			_, err = db.Exec("INSERT INTO pocket values(?, 1);", noteID)
			if err != nil { log.Fatal( err ) }
			return
		} else {
			log.Fatal(err)
		}
	}
	// note is already pocketed & is rank 1 -> do nothing
	if rank == 1 {return}
	
	// note is already pocketed, but not rank 1 -> remove the row, increment everything with a lesser rank to it by 1, add it at rank 1
	_, err := db.Exec("DELETE FROM pocket WHERE note = ?;", noteID)
	if err != nil { log.Fatal( err ) }
	_, err = db.Exec("UPDATE pocket SET rank = rank + 1 WHERE rank < ?;", rank)
	if err != nil { log.Fatal( err ) }
	_, err = db.Exec("INSERT INTO pocket values(?, 1);", noteID)
	if err != nil { log.Fatal( err ) }
}

func getPocket(db *sql.DB) []string {
	noteRows, err := db.Query("SELECT note FROM pocket ORDER BY rank ASC;")
	if err != nil { log.Fatal(err) }
	defer noteRows.Close()

	list := make([]string, 0)

	for noteRows.Next() {
		
		var id string
		if err := noteRows.Scan(&id); err != nil {
			log.Print(err)
		}
		list = append(list, id)
	}
	return list
}

func pushListToPocket(notes []note, db *sql.DB) {

	for i := len(notes) - 1 ; i >= 0; i-- {// Adding vlaues in reverse as that's how the pocke works with searhc priporty (? maybe ?), TODO can using len cause issues when removing elemens from a slice
		pushNoteToPocket(notes[i].id, db)
	}
}

// TODO handle empty pocket
func searchFullPocket(db *sql.DB) []note {
	return searchByIDs(getPocket(db), db)
}

func searchSinglePocket(rank int, db *sql.DB) note {
	list := getPocket(db)
	print(len(list))
	if len(list) >= rank {
		return searchByID(list[rank - 1], db)
	}

	log.Fatal("Rank is greater than length of pocket.")// TODO - This was originally left as a non-fatal log... but I don't remeber why? Was it left to allow searching up-to the max rank? When rank was supposed t hebace somewhat like count?
	var n note
	return n
}

func dropFromPocket(noteID string, db *sql.DB) {
	var rank int
	if err := db.QueryRow("SELECT rank FROM pocket WHERE note = ?;", noteID).Scan(&rank); err != nil {
		if err == sql.ErrNoRows { return } else { log.Fatal(err) }
	}

	_, err := db.Exec("DELETE FROM pocket WHERE note = ?;", noteID)
	if err != nil { log.Fatal( err ) }
	_, err = db.Exec("UPDATE pocket SET rank = rank - 1 WHERE rank > ?;", rank)
	if err != nil { log.Fatal( err ) }

}

func emptyPocket(db *sql.DB) {
	_, err := db.Exec("DELETE FROM pocket;")
	if err != nil { log.Fatal( err ) }
}
