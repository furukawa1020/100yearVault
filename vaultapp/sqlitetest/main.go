package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbFile := "test.db"
	os.Remove(dbFile)

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	CREATE TABLE vaults (id TEXT PRIMARY KEY, title TEXT, state TEXT);
	INSERT INTO vaults (id, title, state) VALUES ('v1', 'My Secret', 'Sealed');
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	rows, err := db.Query("SELECT id, title, state FROM vaults")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, title, state string
		err = rows.Scan(&id, &title, &state)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %s, Title: %s, State: %s\n", id, title, state)
	}

	fmt.Println("SQLite test passed.")
}
