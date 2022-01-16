package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"src/todo"
	"strings"
	"time"
)

/*
REVISION HISTORY
-------- -------
11 Jan 22 -- Started copying out of "Powerful Command-Line Applications in Go" by Ricardo Gerardi
15 Jan 22 -- Modifying the output for the -h flag using the book code.  I don't need -v flag anymore.
             Then added the String method, but that had to be a value receiver to work as like in the book.
             Then added use of TODO_FILENAME environment variable.
*/

const lastModified = "16 Jan 2022"

var todoFilename = "todo.json" // now a var instead of a const so can use environment variable if set.
var todoFileBin = "todo.gob"   // now a var instead of a const so can use environment variable if set.
var fileExists bool

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
		fmt.Printf(" Default file root is todo for todo.json and todo.gob.  TODO_FILENAME environment variable is read.\n")

	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.UserHomeDir is %v.\n", err)
	}

	envValue, ok := os.LookupEnv("TODO_FILENAME")
	if ok {
		todoFilename = envValue + ".json"
		todoFileBin = filepath.Base(envValue) + ".gob"
	}
	if *verboseFlag {
		fmt.Printf(" todoFilename = %s, todoFileBin = %s\n", todoFilename, todoFileBin)
	}

	fullFilenameJson := filepath.Join(homeDir, todoFilename)
	fullFilenameBin := filepath.Join(homeDir, todoFileBin)
	if *verboseFlag {
		fmt.Printf(" fullFilenameJson = %s, fullFilenameBin = %s\n", fullFilenameJson, fullFilenameBin)
	}

	_, err = os.Stat(fullFilenameJson)
	if err != nil {
		//fmt.Fprintf(os.Stderr, " %s got error from os.Stat of %v.\n", fullFilenameJson, err)
		fileExists = false
	} else {
		fileExists = true
	}
	_, err = os.Stat(fullFilenameBin)
	if err != nil {
		//fmt.Fprintf(os.Stderr, " %s got error from os.Stat of %v.\n", fullFilenameBin, err)
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
			fileExists = false
		}
	}

	if *verboseFlag {
		for i, t := range l {
			fmt.Printf(" %d: %s, %t, %s, %s\n", i+1, t.Task, t.Done, t.CreatedAt.Format(time.RFC822), t.CompletedAt.Format(time.RFC822))
		}
		fmt.Println()
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
		if fileExists {
			fmt.Println(l)
		} else {
			fmt.Fprintf(os.Stderr, " Cannot list todo files (%s or %s) as they cannot be found.\n", fullFilenameJson, fullFilenameBin)
		}

		//fmt.Printf("%s", l)   // this does not work.
		//fmt.Print(l.String()) // this works.  But I figured out why it didn't work at first like the book said it should.  See the above comment.
	case *complete > 0:
		if !fileExists {
			fmt.Fprintf(os.Stderr, " Cannot complete todo entries because files (%s or %s) cannot be found.\n", fullFilenameJson, fullFilenameBin)
			os.Exit(1)
		}
		err = l.Complete(*complete)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Item number %d cannot be completed because %v\n", *complete, err)
		}

		err = l.SaveJSON(fullFilenameJson)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in json because %v\n", err)
		}
		err = l.SaveBinary(fullFilenameBin)
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
