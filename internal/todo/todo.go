// internal/todo.go
package todo

import (
	"fmt"
	"time"

	"github.com/alexeyco/simpletable"
)

type item struct {
	ID          int
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

type Todos struct {
	db *DB
}

func NewTodos(db *DB) *Todos {
	return &Todos{db: db}
}

func (t *Todos) Add(task string) error {
	return t.db.AddTodo(task)
}

func (t *Todos) Complete(id int) error {
	return t.db.CompleteTodo(id)
}

func (t *Todos) Delete(id int) error {
	return t.db.DeleteTodo(id)
}

func (t *Todos) List() ([]item, error) {
	return t.db.GetAllTodos()
}

func (t *Todos) Print() error {
	todos, err := t.db.GetAllTodos()
	if err != nil {
		return err
	}

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "ID"},
			{Align: simpletable.AlignCenter, Text: "Task"},
			{Align: simpletable.AlignCenter, Text: "Done"},
			{Align: simpletable.AlignRight, Text: "CreatedAt"},
			{Align: simpletable.AlignRight, Text: "CompletedAt"},
		},
	}

	var cells [][]*simpletable.Cell

	for _, item := range todos {
		task := blue(item.Task)
		done := blue("No")
		if item.Done {
			task = green(fmt.Sprintf("* %s", item.Task))
			done = green("Yes")
		}
		cells = append(cells, []*simpletable.Cell{
			{Text: fmt.Sprintf("%d", item.ID)},
			{Text: task},
			{Text: done},
			{Text: item.CreatedAt.Format(time.RFC822)},
			{Text: item.CompletedAt.Format(time.RFC822)},
		})
	}

	table.Body = &simpletable.Body{Cells: cells}

	table.Footer = &simpletable.Footer{Cells: []*simpletable.Cell{
		{Align: simpletable.AlignCenter, Span: 5, Text: red(fmt.Sprintf("you have %d pending todos", t.CountPending()))},
	}}

	table.SetStyle(simpletable.StyleUnicode)

	table.Println()
	return nil
}

func (t *Todos) CountPending() int {
	todos, err := t.db.GetPendingTodos()
	// TODO: Handle this excpetion better...
	if err != nil {
		return 0
	}
	return len(todos)
}

func (t *Todos) GetStandupTasks(currentTime time.Time) ([]string, time.Time) {
	// Get the current day
	weekday := currentTime.Weekday()
	var lookbackDays int
	if weekday == time.Monday {
		// If it is a Monday, look back 3 day
		lookbackDays = 3
	} else {
		// Any other day, just use 1 day
		lookbackDays = 1
	}
	lookbackDate := currentTime.AddDate(0, 0, -lookbackDays)

	todos, err := t.db.GetCompletedTodos(lookbackDate)
	if err != nil {
		return nil, lookbackDate
	}

	var tasks []string
	for _, item := range todos {
		tasks = append(tasks, item.Task)
	}

	return tasks, lookbackDate
}

func (t *Todos) GetTasks(currentTime time.Time) ([]string, time.Time) {
	todos, err := t.db.GetPendingTodos()
	if err != nil {
		return nil, currentTime
	}

	var tasks []string
	for _, item := range todos {
		tasks = append(tasks, item.Task)
	}

	return tasks, currentTime
}
