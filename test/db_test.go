package todo

import (
	"github.com/JoseTorrado/todo-cli/internal/todo"
	"os"
	"testing"
)

func TestNewDB(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	db, err := todo.NewDB(tempFile.Name())
	if err != nil {
		t.Errorf("NewDB returned the following error: %v", err)
	}
	if db == nil {
		t.Error("NewDB returned nil db!")
	}

	// Closing the database
	if db != nil {
		db.Close()
	}

}
