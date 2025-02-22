package main

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	createQuery := `CREATE TABLE IF NOT EXISTS 
	people 
	(id INTEGER PRIMARY KEY, firstname TEXT, lastname TEXT)`
	database, _ := sql.Open("sqlite3", "./test.db")
	statement, _ := database.Prepare(createQuery)
	statement.Exec()

	insertQuery := `INSERT INTO people
	(firstname, lastname) VALUES (?,?)`
	statement, _ = database.Prepare(insertQuery)
	statement.Exec("Megan", "Wadsworth")

	rows, _ := database.Query("SELECT id, firstname, lastname FROM people")
	var id int
	var firstname string
	var lastname string

	for rows.Next() {
		rows.Scan(&id, &firstname, &lastname)
		fmt.Println(strconv.Itoa(id) + ": " + firstname + " " + lastname)
	}
}
