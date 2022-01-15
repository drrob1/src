package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"src/todo"
	"strings"
)

/*
REVISION HISTORY
-------- -------
11 Jan 22 -- Started copying out of "Powerful Command-Line Applications in Go" by Ricardo Gerardi
15 Jan 22 -- Modifying the output for the -h flag using the book code.  I don't need -v flag anymore.
*/

const lastModified = "15 Jan 2022"

const todoFilename = "todo.json"
const todoFileBin = "todo.gob"

var verboseFlag = flag.Bool("v", false, "Set verbose mode.")
var task = flag.String("task", "", "Task to be added to the ToDo list.")
var complete = flag.Int("complete", 0, "Item to be completed.") // here, 0 means NTD.  That's why we have to start at 1 for item numbers.
var listFlag = flag.Bool("list", false, "List all tasks to the display.")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last modified %s. \n", os.Args[0], lastModified)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *verboseFlag {
		fmt.Printf(" todo last modified %s.  It will display and manage a todo list.\n", lastModified)

	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.UserHomeDir is %v.\n", err)
	}

	fullFilenameJson := filepath.Join(homeDir, todoFilename)
	fullFilenameBin := filepath.Join(homeDir, todoFileBin)

	_, err = os.Stat(fullFilenameJson)
	if err != nil {
		fmt.Fprintf(os.Stderr, " %s got error from os.Stat of %v.\n", fullFilenameJson, err)
	}
	_, err = os.Stat(fullFilenameBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, " %s got error from os.Stat of %v.\n", fullFilenameBin, err)
	}

	l := todo.ListType{}
	err = l.LoadJSON(fullFilenameJson) // if file doesn't exist, this doesn't return an error.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error returned while reading %s is %v\n", fullFilenameJson, err)
		er := l.LoadBinary(fullFilenameBin)
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error returned while reading %s is %v\n", fullFilenameBin, er)
			fmt.Print(" Should I exit? ")
			var ans string
			fmt.Scanln(&ans)
			if strings.HasPrefix(strings.ToLower(ans), "y") {
				os.Exit(1)
			}
			fmt.Println()
		}
	}

	switch {
	case *listFlag:
		/* Replaced by the stringer interface
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
		*/

		// This should invoke the stringer interface from the fmt package.  IE, call the String method I defined in todo.  But it's not working.
		// I kept playing w/ it and I read the docs at golang.org.  I concluded that the stringer interface required a value receiver.  I had
		// followed the book that defined it as a pointer receiver.  So I defined it in todo.go as a value receiver, and it started to work.
		fmt.Println(l)
		//fmt.Printf("%s", l)   // this does not work.
		//fmt.Print(l.String()) // this works.  But I figured out why it didn't work at first like the book said it should.  See the above comment.
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
		err = l.SaveJSON(fullFilenameJson)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in JSON because %v \n", err)
		}
		err = l.SaveBinary(fullFilenameBin)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in binary format because %v \n", err)
		}
	default:
		fmt.Fprintf(os.Stderr, " No valid option was set.\n")
	}
}
