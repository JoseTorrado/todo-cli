package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/JoseTorrado/todo-cli/internal/todo"
)

const (
	todoFileName = ".todos.json"
)

const (
	dbName = "todos.db"
)

func main() {

	// Flexible way to get the same file in the home directory
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return
	}
	// todoFile := filepath.Join(usr.HomeDir, todoFileName)
	dbPath := filepath.Join(usr.HomeDir, ".todo", dbName)

	// Ensure the directory exists
	err = os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// INitialize the DB
	db, err := todo.NewDB(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error intializing the databsae: ", err)
		os.Exit(1)
	}
	defer db.Close()

	// INitialize the Schema
	if err := db.InitSchema(); err != nil {
		fmt.Fprintln(os.Stderr, "Error intializing db schema: ", err)
		os.Exit(1)
	}

	// Create a new Todos instance
	todos := todo.NewTodos(db)

	add := flag.Bool("add", false, "Add a new todo")
	complete := flag.Int("done", 0, "Mark a todo as Completed")
	del := flag.Int("rm", 0, "Delete a todo")
	list := flag.Bool("ls", false, "List all the todos")
	standup := flag.Bool("standup", false, "Print all tasks completed yesterday")
	today := flag.Bool("today", false, "Print all tasks leftover for today")

	flag.Parse()

	switch {
	case *add:
		task, err := getInput(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		if err := todos.Add(task); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

	case *complete > 0:
		err := todos.Complete(*complete)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

	case *del > 0:
		err := todos.Delete(*del)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

	case *list:
		if err := todos.Print(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

	case *standup:
		tasks, lookbackDate := todos.GetStandupTasks(time.Now())

		// Print the lookback date
		fmt.Printf("%s:\n", lookbackDate.Format("2006-01-02"))

		// Loop through the tasks and print them
		if len(tasks) == 0 {
			fmt.Println("No tasks recorded.")
		} else {
			for _, task := range tasks {
				fmt.Printf("* %s\n", task)
			}
		}

	case *today:
		tasks, currentDate := todos.GetTasks(time.Now())

		// Print the lookback date
		fmt.Printf("%s:\n", currentDate.Format("2006-01-02"))

		// Loop through the tasks and print them
		if len(tasks) == 0 {
			fmt.Println("No tasks recorded.")
		} else {
			for _, task := range tasks {
				fmt.Printf("* %s\n", task)
			}
		}

	default:
		fmt.Fprintln(os.Stdout, "invalid command passed")
		os.Exit(0)
	}
}

// getting text input for a Todo name
func getInput(r io.Reader, args ...string) (string, error) {

	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}

	text := scanner.Text()

	if len(text) == 0 {
		return "", nil
	}

	return text, nil
}
