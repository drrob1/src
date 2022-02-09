package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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
16 Jan 22 -- Added stdin as a source.  And changed name of string task flag to a boolean add flag.
17 Jan 22 -- Added default actions if no switch is provided.  If there are arguments then add as a task, if not list tasks.
 8 Feb 22 -- Will show a timestamp of adding a task, done by updating the stringer method in todo.go.  I changed how the
               filename is constructed.  I am considering adding another environment variable, called TODO_PREFIX to more easily cover the networking prefix.
*/

const lastModified = "9 Feb 2022"

var todoFilename = "todo.json" // now a var instead of a const so can use environment variable if set.
var todoFileBin = "todo.gob"   // now a var instead of a const so can use environment variable if set.
var prefix string
var fileExists bool

var verboseFlag = flag.Bool("v", false, "Set verbose mode.")

//                                                                                  var task = flag.String("task", "", "Task to be added to the ToDo list.")
var add = flag.Bool("add", false, "Add task to the ToDo list.")
var complete = flag.Int("complete", 0, "Item to be completed.") // here, 0 means NTD.  That's why we have to start at 1 for item numbers.
var listFlag = flag.Bool("list", false, "List all tasks to the display.")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last modified %s. \n", os.Args[0], lastModified)
		fmt.Fprintf(flag.CommandLine.Output(), "TODO_PREFIX and TODO_FILENAME are the environment variables used.  Do not use an extension for TODO_FILENAME")
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *verboseFlag {
		fmt.Printf(" todo last modified %s.  It will display and manage a todo list.\n", lastModified)
		fmt.Printf(" Default filename root is todo for todo.json and todo.gob.  TODO_FILENAME environment variable is read, and should not have an extension.\n")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.UserHomeDir is %v.\n", err)
	}

	var fullFilenameJson, fullFilenameBin string
	envFN, ok := os.LookupEnv("TODO_FILENAME")
	if ok {
		todoFilename = envFN + ".json"
		todoFileBin = filepath.Base(envFN) + ".gob"
	}
	if *verboseFlag {
		fmt.Printf(" todoFilename = %s, todoFileBin = %s\n", todoFilename, todoFileBin)
	}

	prefix, ok = os.LookupEnv("TODO_PREFIX")
	if ok {
		fullFilenameJson = filepath.Join(prefix, todoFilename)
		fullFilenameBin = filepath.Join(prefix, todoFileBin)
	} else {
		fullFilenameJson = filepath.Join(homeDir, todoFilename)
		fullFilenameBin = filepath.Join(homeDir, todoFileBin)
	}

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
	case *add:
		task, err := getTask(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		l.Add(task)
		err = l.SaveJSON(fullFilenameJson)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in JSON because %v \n", err)
		}
		err = l.SaveBinary(fullFilenameBin)
		if err != nil {
			fmt.Fprintf(os.Stderr, " List could not be saved in binary format because %v \n", err)
		}
	default: // task add
		if flag.NArg() > 0 {
			tsk := strings.Join(flag.Args(), " ")
			l.Add(tsk)
			err = l.SaveJSON(fullFilenameJson)
			if err != nil {
				fmt.Fprintf(os.Stderr, " List could not be saved in JSON because %v \n", err)
			}
			err = l.SaveBinary(fullFilenameBin)
			if err != nil {
				fmt.Fprintf(os.Stderr, " List could not be saved in binary format because %v \n", err)
			}
		} else {
			if fileExists {
				fmt.Println(l)
			} else {
				fmt.Fprintf(os.Stderr, " Cannot list todo files (%s or %s) as they cannot be found.\n", fullFilenameJson, fullFilenameBin)
			}
		}
	}
}

func getTask(r io.Reader, args ...string) (string, error) { // decides where to get task string, from args or stdin.
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	scnr := bufio.NewScanner(r)
	scnr.Scan()
	if err := scnr.Err(); err != nil {
		return "", err
	}

	if len(scnr.Text()) == 0 {
		return "", fmt.Errorf("Task to add cannot be blank.")
	}

	return scnr.Text(), nil
}
