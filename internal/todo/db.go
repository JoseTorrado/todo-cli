package todo

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
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

func (db *DB) DeleteTodo(id int) error {
	_, err := db.Exec(`
		DELETE FROM todos WHERE id = ?
		`, id)
	return err
}

// THis helper function allows to pass any datatype into the query parameters by assigning it the interface type
func (db *DB) scanTodos(query string, args ...interface{}) ([]item, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []item
	for rows.Next() {
		var i item
		var completedAt sql.NullTime
		err := rows.Scan(&i.ID, &i.Task, &i.Done, &i.CreatedAt, &completedAt)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			i.CompletedAt = completedAt.Time
		}
		todos = append(todos, i)
	}
	return todos, nil

}

func (db *DB) GetAllTodos() ([]item, error) {
	return db.scanTodos(`
		SELECT
				id,
				task,
				done,
				created_at,
				completed_at
		FROM
				todos;
		`)
}

func (db *DB) GetCompletedTodos(since time.Time) ([]item, error) {
	return db.scanTodos(`
		SELECT 
				id, 
				task,
				done,
				created_at,
				completed_at
		FROM 
				todos
		WHERE
				done = 1
				AND completed_at > ?;
		`, since)
}

func (db *DB) GetPendingTodos() ([]item, error) {
	return db.scanTodos(`
		SELECT
				id,
				task,
				done,
				created_at,
				completed_at
		FROM
				todos 
		WHERE 
				done = 0;
		`)
}

func (db *DB) GetRecentOrPendingTodos(since time.Time) ([]item, error) {
	return db.scanTodos(`
		SELECT 
				id, 
				task,
				done,
				created_at,
				completed_at
		FROM 
				todos
		WHERE
				done = 1
				AND completed_at > ?;
		`, since)
}
