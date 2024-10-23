package test

import (
	"os"
	"testing"
	"time"

	"github.com/JoseTorrado/todo-cli/internal/todo"
)

func TestAdd(t *testing.T) {
	// Arrange
	var todos todo.Todos
	task := "Test Task"

	// Action
	todos.Add(task)

	// Assert
	if len(todos) != 1 {
		t.Errorf("expected 1 todo but got %d", len(todos))
	}

	addedTask := todos[0]

	if addedTask.Task != task {
		t.Errorf("Expected task to be %q but got %q", task, addedTask.Task)
	}

	if addedTask.Done != false {
		t.Errorf("Expected Done to be fasle but got %v", addedTask.Done)
	}

	if addedTask.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be set but got zero value")
	}

	if !addedTask.CompletedAt.IsZero() {
		t.Errorf("Expected CompletedAt to be set but got zero value")
	}
}

func TestComplete(t *testing.T) {
	//Arrange
	todos := todo.Todos{
		{Task: "Task 1", Done: false, CreatedAt: time.Now(), CompletedAt: time.Time{}},
		{Task: "Task 2", Done: false, CreatedAt: time.Now(), CompletedAt: time.Time{}},
	}

	// Action
	err := todos.Complete(1)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}

	if !todos[0].Done {
		t.Errorf("Expected Done to be true but got %v", todos[0].Done)
	}

	if todos[0].CompletedAt.IsZero() {
		t.Error("Expected CompletedAt to be set byt for a zero value")
	}

	// Testing out of bounds
	err = todos.Complete(3)
	if err == nil {
		t.Error("Expected Error but got nil")
	}

	if err.Error() != "Invalid Index" {
		t.Errorf("Epected error bessage 'Invalid Index' but got %v", err.Error())
	}
}

func TestDelete(t *testing.T) {
	// Arrange
	todos := todo.Todos{
		{Task: "Task 1", Done: false, CreatedAt: time.Now(), CompletedAt: time.Time{}},
		{Task: "Task 2", Done: false, CreatedAt: time.Now(), CompletedAt: time.Time{}},
	}

	// Action
	err := todos.Delete(1)

	// Assert
	if len(todos) != 1 {
		t.Errorf("expected 1 todo after deletion but got %d", len(todos))
	}

	// Testing out of bounds
	err = todos.Delete(2)
	if err == nil {
		t.Error("Expected Error but got nil")
	}

	if err.Error() != "Invalid Index" {
		t.Errorf("Epected error bessage 'Invalid Index' but got %v", err.Error())
	}
}

func TestLoad(t *testing.T) {
	// Arrange
	// Create a temp file to read from
	tempFile, err := os.CreateTemp("", "todos.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Schedule removal of temp file after test runs
	defer os.Remove(tempFile.Name())

	// Write Data to JSON
	jsonData := `[{"Task":"Task 1", "Done":false},{"Task":"Task 2", "Done":true}]`
	if _, err := tempFile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Close file after write
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Action
	var extractedTodos todo.Todos
	err = extractedTodos.Load(tempFile.Name())

	// Assert
	if len(extractedTodos) != 2 {
		t.Errorf("Expected 2 todos but got %d", len(extractedTodos))
	}

	// Verify the todos were extracted properly
	if extractedTodos[0].Task != "Task 1" || extractedTodos[1].Task != "Task 2" {
		t.Errorf("unexpected tasks: %+v", extractedTodos)
	}
	if extractedTodos[0].Done != false || extractedTodos[1].Done != true {
		t.Errorf("unexpected done statuses: %+v", extractedTodos)
	}
}

func TestGetStandupTasks_Monday(t *testing.T) {
	// Mocking the current time to simulate a Monday
	mockMonday := time.Date(2024, 9, 16, 0, 0, 0, 0, time.UTC) // A Monday

	// Create tasks with different completion times
	todos := todo.Todos{
		{Task: "Task 1", Done: true, CompletedAt: time.Date(2024, 9, 13, 14, 0, 0, 0, time.UTC)},  // Completed on Friday
		{Task: "Task 2", Done: true, CompletedAt: time.Date(2024, 9, 14, 12, 0, 0, 0, time.UTC)},  // Completed on Saturday
		{Task: "Task 3", Done: false, CompletedAt: time.Date(2024, 9, 13, 11, 0, 0, 0, time.UTC)}, // Not Done
	}

	// Run the GetStandupTasks function with mockMonday
	tasks, lookbackDate := todos.GetStandupTasks(mockMonday)

	// Define the expected output
	expectedTasks := []string{"Task 1"}
	expectedLookbackDate := time.Date(2024, 9, 13, 0, 0, 0, 0, time.UTC) // Friday

	// Validate the lookback date
	if !lookbackDate.Equal(expectedLookbackDate) {
		t.Errorf("Expected lookback date %v but got %v", expectedLookbackDate, lookbackDate)
	}

	// Validate the tasks
	if len(tasks) != len(expectedTasks) {
		t.Errorf("Expected %d tasks but got %d", len(expectedTasks), len(tasks))
	}

	for i, task := range tasks {
		if task != expectedTasks[i] {
			t.Errorf("Expected task %q but got %q", expectedTasks[i], task)
		}
	}
}

func TestGetStandupTasks_RegularDay(t *testing.T) {
	// Mocking the current time to simulate a Wednesday
	mockWednesday := time.Date(2024, 9, 18, 0, 0, 0, 0, time.UTC) // A Wednesday

	// Create tasks with different completion times
	todos := todo.Todos{
		{Task: "Task 1", Done: true, CompletedAt: time.Date(2024, 9, 17, 14, 0, 0, 0, time.UTC)},  // Completed on Tuesday
		{Task: "Task 2", Done: true, CompletedAt: time.Date(2024, 9, 16, 12, 0, 0, 0, time.UTC)},  // Completed on Monday
		{Task: "Task 3", Done: false, CompletedAt: time.Date(2024, 9, 17, 11, 0, 0, 0, time.UTC)}, // Not Done
	}

	// Run the GetStandupTasks function with mockWednesday
	tasks, lookbackDate := todos.GetStandupTasks(mockWednesday)

	// Define the expected output
	expectedTasks := []string{"Task 1"}
	expectedLookbackDate := time.Date(2024, 9, 17, 0, 0, 0, 0, time.UTC) // Tuesday

	// Validate the lookback date
	if !lookbackDate.Equal(expectedLookbackDate) {
		t.Errorf("Expected lookback date %v but got %v", expectedLookbackDate, lookbackDate)
	}

	// Validate the tasks
	if len(tasks) != len(expectedTasks) {
		t.Errorf("Expected %d tasks but got %d", len(expectedTasks), len(tasks))
	}

	for i, task := range tasks {
		if task != expectedTasks[i] {
			t.Errorf("Expected task %q but got %q", expectedTasks[i], task)
		}
	}
}

func TestGetTaks(t *testing.T) {
	// Mocking the current time to simulate a Wednesday
	mockWednesday := time.Date(2024, 9, 18, 0, 0, 0, 0, time.UTC) // A Wednesday

	// Create tasks with different completion times
	todos := todo.Todos{
		{Task: "Task 1", Done: true, CompletedAt: time.Date(2024, 9, 17, 14, 0, 0, 0, time.UTC)},  // Completed on Tuesday
		{Task: "Task 2", Done: false, CompletedAt: time.Date(2024, 9, 16, 12, 0, 0, 0, time.UTC)}, // Not Done
		{Task: "Task 3", Done: false, CompletedAt: time.Date(2024, 9, 17, 11, 0, 0, 0, time.UTC)}, // Not Done
	}

	// Run the GetStandupTasks function with mockWednesday
	tasks, lookbackDate := todos.GetTasks(mockWednesday)

	// Define the expected output
	expectedTasks := []string{"Task 2", "Task 3"}
	expectedLookbackDate := time.Date(2024, 9, 18, 0, 0, 0, 0, time.UTC) // Same as input

	// Validate the lookback date
	if !lookbackDate.Equal(expectedLookbackDate) {
		t.Errorf("Expected lookback date %v but got %v", expectedLookbackDate, lookbackDate)
	}

	// Validate the tasks
	if len(tasks) != len(expectedTasks) {
		t.Errorf("Expected %d tasks but got %d", len(expectedTasks), len(tasks))
	}

	for i, task := range tasks {
		if task != expectedTasks[i] {
			t.Errorf("Expected task %q but got %q", expectedTasks[i], task)
		}
	}
}
