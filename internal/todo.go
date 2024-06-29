// internal/todo.go
package todo

import (
	"errors"
	"time"
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
