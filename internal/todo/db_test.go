package todo

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Test Helper ---
func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := tempDir + "/test_todo.db"

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.InitSchema()
	if err != nil {
		db.Close()
		t.Fatalf("Failed to initializ schema: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup

}

// --- Actual Tests ---

func TestNewDB(t *testing.T) {
	t.Run("Successful creaton (temp-file)", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test_newdb.db")

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("NewDB(%q) failed: %v", dbPath, err)
		}

		if db == nil {
			t.Fatalf("NewDB(%q) returned nil but no error", dbPath)
		}
		defer db.Close()

		// Ping to ensure conenction is alive
		if err := db.Ping(); err != nil {
			t.Errorf("db.Ping() failed for NewDB (file): %v", err)
		}
	})

	t.Run("Failure case (invalid path)", func(t *testing.T) {
		_, err := NewDB("/?invalidpath")
		if err == nil {
			t.Error("NewDB(\"\") succeeded unexpectedly, expected an error!")
		}
	})
}

func TestInitSchema(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Idempotency check
	err := db.InitSchema()
	if err != nil {
		t.Fatalf("Running Init Schema a second time failed: %v", err)
	}

	var tableName string
	queryErr := db.DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='todos'").Scan(&tableName)

	if queryErr != nil {
		if queryErr == sql.ErrNoRows {
			t.Fatal("InitSchema did not create the todos table")
		}
		t.Fatalf("Fialed to query 'todos' from sqlite_master: %v", queryErr)
		if tableName != "todos" {
			t.Errorf("Expected table name 'todos', but found '%s'", tableName)
		}
	}

}

func TestAddTodo(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	task := "Buy milk"
	// Call AddTodo method directly
	err := db.AddTodo(task)
	if err != nil {
		t.Fatalf("AddTodo failed: %v", err)
	}

	// Verify the todo was added correctly
	var id int
	var taskRead string
	var done bool
	var createdAt time.Time
	var completedAt sql.NullTime // Use sql.NullTime for potentially NULL columns

	query := "SELECT id, task, done, created_at, completed_at FROM todos WHERE task = ?"
	// Query using the embedded *sql.DB
	row := db.DB.QueryRow(query, task)
	err = row.Scan(&id, &taskRead, &done, &createdAt, &completedAt)

	if err != nil {
		t.Fatalf("Failed to query back added todo: %v", err)
	}

	if taskRead != task {
		t.Errorf("Expected task '%s', got '%s'", task, taskRead)
	}
	if done != false {
		t.Errorf("Expected done to be false, got %v", done)
	}
	// Check if timestamp is recent (within a reasonable threshold)
	if time.Since(createdAt) > 5*time.Second {
		t.Errorf("Expected created_at to be recent, got %v", createdAt)
	}
	if completedAt.Valid {
		t.Errorf("Expected completed_at to be NULL (invalid), but it was valid: %v", completedAt.Time)
	}
	if id <= 0 {
		t.Errorf("Expected a positive ID, got %d", id)
	}
}

// Helper to add a task and get its ID for other tests
// Takes *DB (local type)
func addTestTask(t *testing.T, db *DB, task string) int {
	t.Helper()
	err := db.AddTodo(task)
	if err != nil {
		t.Fatalf("Helper addTestTask failed to add '%s': %v", task, err)
	}
	var id int
	// Query using the embedded *sql.DB
	err = db.DB.QueryRow("SELECT id FROM todos WHERE task = ? ORDER BY id DESC LIMIT 1", task).Scan(&id)
	if err != nil {
		t.Fatalf("Helper addTestTask failed to get ID for '%s': %v", task, err)
	}
	return id
}

func TestCompleteTodo(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	task := "Walk the dog"
	id := addTestTask(t, db, task) // Use helper

	// Call CompleteTodo method directly
	err := db.CompleteTodo(id)
	if err != nil {
		t.Fatalf("CompleteTodo failed: %v", err)
	}

	// Verify completion status
	var done bool
	var completedAt sql.NullTime
	query := "SELECT done, completed_at FROM todos WHERE id = ?"
	// Query using the embedded *sql.DB
	err = db.DB.QueryRow(query, id).Scan(&done, &completedAt)
	if err != nil {
		t.Fatalf("Failed to query back completed todo: %v", err)
	}

	if !done {
		t.Errorf("Expected done to be true after CompleteTodo, got false")
	}
	if !completedAt.Valid {
		t.Error("Expected completed_at to be valid (not NULL) after CompleteTodo")
	} else if time.Since(completedAt.Time) > 5*time.Second { // Check if timestamp is recent
		t.Errorf("Expected completed_at to be recent, got %v", completedAt.Time)
	}

	// Test completing a non-existent ID (should not error)
	nonExistentID := 99999
	err = db.CompleteTodo(nonExistentID)
	if err != nil {
		t.Errorf("CompleteTodo for non-existent ID %d failed unexpectedly: %v", nonExistentID, err)
	}
}

func TestDeleteTodo(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	task := "Clean the house"
	id := addTestTask(t, db, task) // Use helper

	// Call DeleteTodo method directly
	err := db.DeleteTodo(id)
	if err != nil {
		t.Fatalf("DeleteTodo failed: %v", err)
	}

	// Verify deletion
	var count int
	query := "SELECT COUNT(*) FROM todos WHERE id = ?"
	// Query using the embedded *sql.DB
	err = db.DB.QueryRow(query, id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count after deleting todo: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count of todo with id %d to be 0 after deletion, got %d", id, count)
	}

	// Test deleting a non-existent ID (should not error)
	nonExistentID := 99998
	err = db.DeleteTodo(nonExistentID)
	if err != nil {
		t.Errorf("DeleteTodo for non-existent ID %d failed unexpectedly: %v", nonExistentID, err)
	}
}

// TestGetAllTodos expects []item (local type)
func TestGetAllTodos(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Empty database", func(t *testing.T) {
		// Call GetAllTodos method directly
		todos, err := db.GetAllTodos()
		if err != nil {
			t.Fatalf("GetAllTodos on empty DB failed: %v", err)
		}
		// Expect []item (local type)
		if len(todos) != 0 {
			t.Errorf("Expected 0 todos on empty DB, got %d", len(todos))
		}
	})

	t.Run("With multiple todos", func(t *testing.T) {
		task1 := "Task A"
		task2 := "Task B"
		id1 := addTestTask(t, db, task1)
		id2 := addTestTask(t, db, task2)

		// Call GetAllTodos method directly
		todos, err := db.GetAllTodos()
		if err != nil {
			t.Fatalf("GetAllTodos failed: %v", err)
		}
		if len(todos) != 2 {
			t.Fatalf("Expected 2 todos, got %d", len(todos))
		}

		// Check if both tasks are present
		foundTask1 := false
		foundTask2 := false
		for _, todoItem := range todos { // todoItem is of type item
			// Access fields directly (assuming they are defined in the item struct)
			if todoItem.ID == id1 && todoItem.Task == task1 {
				foundTask1 = true
			}
			if todoItem.ID == id2 && todoItem.Task == task2 {
				foundTask2 = true
			}
		}

		if !foundTask1 {
			t.Errorf("Did not find task '%s' (ID: %d) in GetAllTodos result", task1, id1)
		}
		if !foundTask2 {
			t.Errorf("Did not find task '%s' (ID: %d) in GetAllTodos result", task2, id2)
		}
	})
}

func TestGetPendingTodos(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Empty database", func(t *testing.T) {
		// Call GetPendingTodos method directly
		todos, err := db.GetPendingTodos()
		if err != nil {
			t.Fatalf("GetPendingTodos on empty DB failed: %v", err)
		}
		if len(todos) != 0 {
			t.Errorf("Expected 0 pending todos on empty DB, got %d", len(todos))
		}
	})

	t.Run("With mixed todos", func(t *testing.T) {
		pendingTask1 := "Pending 1"
		pendingTask2 := "Pending 2"
		completedTask := "Completed 1"

		idPending1 := addTestTask(t, db, pendingTask1)
		idPending2 := addTestTask(t, db, pendingTask2)
		idCompleted := addTestTask(t, db, completedTask)

		// Complete one task
		err := db.CompleteTodo(idCompleted)
		if err != nil {
			t.Fatalf("Failed to complete task for test setup: %v", err)
		}

		// Get pending todos
		pendingTodos, err := db.GetPendingTodos()
		if err != nil {
			t.Fatalf("GetPendingTodos failed: %v", err)
		}
		if len(pendingTodos) != 2 {
			t.Fatalf("Expected 2 pending todos, got %d", len(pendingTodos))
		}

		// Verify the correct todos are returned and are marked as not done
		foundPending1 := false
		foundPending2 := false
		for _, todoItem := range pendingTodos { // todoItem is of type item
			if todoItem.ID == idPending1 && todoItem.Task == pendingTask1 {
				foundPending1 = true
				if todoItem.Done { // Access Done field
					t.Errorf("Pending task '%s' was incorrectly marked as done", pendingTask1)
				}
			}
			if todoItem.ID == idPending2 && todoItem.Task == pendingTask2 {
				foundPending2 = true
				if todoItem.Done {
					t.Errorf("Pending task '%s' was incorrectly marked as done", pendingTask2)
				}
			}
			if todoItem.ID == idCompleted {
				t.Errorf("Completed task '%s' was incorrectly returned by GetPendingTodos", completedTask)
			}
		}

		if !foundPending1 {
			t.Errorf("Did not find pending task '%s' (ID: %d) in GetPendingTodos result", pendingTask1, idPending1)
		}
		if !foundPending2 {
			t.Errorf("Did not find pending task '%s' (ID: %d) in GetPendingTodos result", pendingTask2, idPending2)
		}
	})
}

func TestGetCompletedTodos(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add tasks
	taskPending := "Still Pending"
	taskCompEarly := "Completed Early"
	taskCompLate1 := "Completed Late 1"
	taskCompLate2 := "Completed Late 2"

	addTestTask(t, db, taskPending)
	idEarly := addTestTask(t, db, taskCompEarly)
	idLate1 := addTestTask(t, db, taskCompLate1)
	idLate2 := addTestTask(t, db, taskCompLate2)

	// Complete tasks at different times
	err := db.CompleteTodo(idEarly)
	if err != nil {
		t.Fatalf("Setup: CompleteTodo failed for idEarly: %v", err)
	}

	// Ensure a time difference for the 'since' parameter
	time.Sleep(50 * time.Millisecond) // Small sleep to ensure time progresses
	timeMid := time.Now()
	time.Sleep(50 * time.Millisecond) // Small sleep to ensure time progresses

	err = db.CompleteTodo(idLate1)
	if err != nil {
		t.Fatalf("Setup: CompleteTodo failed for idLate1: %v", err)
	}
	err = db.CompleteTodo(idLate2)
	if err != nil {
		t.Fatalf("Setup: CompleteTodo failed for idLate2: %v", err)
	}
	timeAfterLate := time.Now()

	t.Run("Get completed since before middle time", func(t *testing.T) {
		// Call GetCompletedTodos method directly
		completedTodos, err := db.GetCompletedTodos(timeMid)
		if err != nil {
			t.Fatalf("GetCompletedTodos failed: %v", err)
		}

		if len(completedTodos) != 2 {
			t.Fatalf("Expected 2 completed todos since timeMid, got %d", len(completedTodos))
		}

		// Verify the correct todos are returned and marked done
		foundLate1 := false
		foundLate2 := false
		for _, todoItem := range completedTodos { // todoItem is of type item
			if todoItem.ID == idLate1 && todoItem.Task == taskCompLate1 {
				foundLate1 = true
				if !todoItem.Done {
					t.Errorf("Task '%s' should be done", taskCompLate1)
				}
				// Check completion time is after 'since' (using CompletedAt field)
				if !todoItem.CompletedAt.After(timeMid) {
					t.Errorf("Task '%s' completion time %v not after timeMid %v", taskCompLate1, todoItem.CompletedAt, timeMid)
				}
			}
			if todoItem.ID == idLate2 && todoItem.Task == taskCompLate2 {
				foundLate2 = true
				if !todoItem.Done {
					t.Errorf("Task '%s' should be done", taskCompLate2)
				}
				if !todoItem.CompletedAt.After(timeMid) {
					t.Errorf("Task '%s' completion time %v not after timeMid %v", taskCompLate2, todoItem.CompletedAt, timeMid)
				}
			}
			if todoItem.ID == idEarly {
				t.Errorf("Early completed task '%s' was incorrectly returned", taskCompEarly)
			}
		}
		if !foundLate1 {
			t.Errorf("Did not find late task 1")
		}
		if !foundLate2 {
			t.Errorf("Did not find late task 2")
		}
	})

	t.Run("Get completed since after all completions", func(t *testing.T) {
		completedTodos, err := db.GetCompletedTodos(timeAfterLate)
		if err != nil {
			t.Fatalf("GetCompletedTodos failed: %v", err)
		}
		if len(completedTodos) != 0 {
			t.Errorf("Expected 0 completed todos since timeAfterLate, got %d", len(completedTodos))
		}
	})

}

// --- End of Tests ---
