package main

import (
	"flag"
	"fmt"
	"os"
	"src/todo"
	"strings"
)

const todoFilename = "todo.json"
const todoFileBin = "todo.gob"
const lastModified = "14 Jan 2022"

var verboseFlag = flag.Bool("v", false, "Set verbose mode.")
var task = flag.String("task", "", "Task to be added to the ToDo list.")
var complete = flag.Int("complete", 0, "Item to be completed.") // here, 0 means NTD.  That's why we have to start at 1 for item numbers.
var listFlag = flag.Bool("list", false, "List all tasks to the display.")

func main() {
	flag.Parse()
	if *verboseFlag {
		fmt.Printf(" todo last modified %s.  It will display and manage a todo list.\n", lastModified)

	}

	l := todo.ListType{}
	err := l.LoadJSON(todoFilename) // if file doesn't exist, this does not return an error.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error returned while reading %s is %v\n", todoFilename, err)
		fmt.Print(" Should I exit? ")
		var ans string
		fmt.Scanln(&ans)
		if strings.HasPrefix(strings.ToLower(ans), "y") {
			os.Exit(1)
		}
	}

	switch {
	case *listFlag:
		for _, item := range l {
			if !item.Done {
				fmt.Printf(" Not done: %s\n", item.Task)
			}
		}
		fmt.Println()
		for _, item := range l {
			if item.Done {
				fmt.Printf(" Done: %s was completed on %s\n", item.Task, item.CompletedAt.Format("Jan-02-2006 15:04:05"))
			}
		}
	case *complete > 0:
		err = l.Complete(*complete)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Item number %d cannot be completed because %v\n", *complete, err)
		}

		err = l.SaveJSON(todoFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in json because %v\n", err)
		}
		err = l.SaveBinary(todoFileBin)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in binary format because %v\n", err)
		}
	case *task != "":
		l.Add(*task)
		err = l.SaveJSON(todoFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in JSON because %v \n", err)
		}
		err = l.SaveBinary(todoFileBin)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in binary format because %v \n", err)
		}
	default:
		fmt.Fprintf(os.Stderr, " No valid option was set.\n")
	}
}
