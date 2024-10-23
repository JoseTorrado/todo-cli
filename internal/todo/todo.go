// internal/todo.go
package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alexeyco/simpletable"
)

type item struct {
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

type Todos []item

func (t *Todos) Add(task string) {
	todo := item{
		Task:        task,
		Done:        false,
		CreatedAt:   time.Now(),
		CompletedAt: time.Time{},
	}

	*t = append(*t, todo)
}

func (t *Todos) Complete(index int) error {
	ls := *t
	// handling errors
	if index <= 0 || index > len(ls) {
		return errors.New("Invalid Index")
	}

	ls[index-1].CompletedAt = time.Now()
	ls[index-1].Done = true

	return nil

}

func (t *Todos) Delete(index int) error {
	ls := *t
	// handling errors
	if index <= 0 || index > len(ls) {
		return errors.New("Invalid Index")
	}

	*t = append(ls[:index-1], ls[index:]...) // ... unpacks the slice to be appended so the append function can add them to the first slice

	return nil
}

func (t *Todos) Load(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(file) == 0 {
		return nil
	}

	err = json.Unmarshal(file, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *Todos) Save(filename string) error {

	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (t *Todos) Print() {

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "id"},
			{Align: simpletable.AlignCenter, Text: "Task"},
			{Align: simpletable.AlignCenter, Text: "Done"},
			{Align: simpletable.AlignRight, Text: "CreatedAt"},
			{Align: simpletable.AlignRight, Text: "CompletedAt"},
		},
	}

	var cells [][]*simpletable.Cell

	for index, item := range *t {
		index++
		if !item.Done || (item.Done && item.CompletedAt.After(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location()))) {
			task := blue(item.Task)
			done := blue("No")
			if item.Done {
				task = green(fmt.Sprintf("* %s", item.Task))
				done = green("Yes")
			}
			cells = append(cells, *&[]*simpletable.Cell{
				{Text: fmt.Sprintf("%d", index)},
				{Text: task},
				{Text: done},
				{Text: item.CreatedAt.Format(time.RFC822)},
				{Text: item.CompletedAt.Format(time.RFC822)},
			})
		}
	}

	table.Body = &simpletable.Body{Cells: cells}

	table.Footer = &simpletable.Footer{Cells: []*simpletable.Cell{
		{Align: simpletable.AlignCenter, Span: 5, Text: red(fmt.Sprintf("you have %d pending todos", t.CountPending()))},
	}}

	table.SetStyle(simpletable.StyleUnicode)

	table.Println()
}

func (t *Todos) CountPending() int {
	total := 0
	for _, item := range *t {
		if !item.Done {
			total++
		}
	}
	return total
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
	lookbackStart := time.Date(lookbackDate.Year(), lookbackDate.Month(), lookbackDate.Day(), 0, 0, 0, 0, lookbackDate.Location())
	lookbackEnd := lookbackStart.AddDate(0, 0, 1).Add(-time.Nanosecond)

	var tasks []string
	for _, item := range *t {
		if item.Done && item.CompletedAt.After(lookbackStart) && item.CompletedAt.Before(lookbackEnd) {
			tasks = append(tasks, item.Task)
		}
	}

	return tasks, lookbackDate
}

func (t *Todos) GetTasks(currentTime time.Time) ([]string, time.Time) {

	var tasks []string
	for _, item := range *t {
		if !item.Done {
			tasks = append(tasks, item.Task)
		}
	}

	return tasks, currentTime
}
