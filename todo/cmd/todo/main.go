package main

import (
	"fmt"
	"os"
	"src/todo"
	"strings"
)

const todoFilename = "todo.json"
const todoFileBin = "todo.gob"
const lastModified = "12 Jan 2022"

func main() {
	//fmt.Printf(" todo last modified %s.  It will display and manage a todo list.\n", lastModified)
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
	case len(os.Args) == 1:
		str := l.List() // if no params, just list the tasks
		for _, s := range str {
			fmt.Println(s)
		}
	default:
		item := strings.Join(os.Args[1:], " ")
		l.Add(item)
		err = l.SaveJSON(todoFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error returned from SaveJSON is %v \n", err)
		}
		err = l.SaveBinary(todoFileBin)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error returned from SaveBinary is %v \n", err)
		}
	}
}
