package todo

import (
	"database/sql"
	"os"
	"time"
)

type DB struct {
	*sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	// We open the db and return the error if one rises
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// This is common Go idiom for error checking
	// It basically assigns the if AFTER the error check, but its a concise way to write
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// If everythign goes well, returnt eh DB object and null fro error
	return &DB{DB: db}, nil
}

// Creating the Schema of our DB
func (db *DB) InitSchema() error {
	_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS todos (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					task TEXT NOT NULL,
					done BOOLEAN NOT NULL DEFAULT 0,
					created_at DATETIME NOT NULL,
					completed_at DATETIME
			)
		`)
	return err
}

func (db *DB) AddTodo(task string) error {
	_, err := db.Exec(`
				INSERT INTO todos
				(task, created_at) VALUES (?, ?)
		`, task, time.Now())

	return err
}

func (db *DB) CompleteTodo(id int) error {
	_, err := db.Exec(`
			UPDATE todos 
			SET done = 1, completed_at = ?
			WHERE id = ?
		`, time.Now(), id)

	return err
}
