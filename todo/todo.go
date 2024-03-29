package todo

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

const lastModified = "Feb 27, 2022"

/*
REVISION HISTORY
-------- -------
10 Jan 22 -- Started writing this code from "Powerful Command-Line Applications in Go" by Ricardo Gerardi
15 Jan 22 -- Changed behavior of the file read to return the error if file was not found.
               Then added the String() method using value receiver as fmt package requires that to work.
 8 Feb 22 -- Will show the created at field in the Stringer method.
26 Feb 22 -- Added GetString to be able to list separately the completed vs the not completed tasks.
 1 Apr 23 -- StaticCheck found a few issues.
*/

type item struct {
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

type ListType []item // the book calls this type List.  I think that's too vague so I changed it.

func (l *ListType) Add(task string) { // note that this is a pointer receiver
	tsk := item{
		Task:        task,
		Done:        false,
		CreatedAt:   time.Now(),
		CompletedAt: time.Time{}, // note that this is a zeroed value
	}
	*l = append(*l, tsk)
}

func (l *ListType) Complete(i int) error { // Will use 1 origin reference so i = 1 is first item
	ls := *l
	if i <= 0 || i > len(ls) {
		return fmt.Errorf("item %d does not exist", i)
	}

	ls[i-1].Done = true // remember i is a 1 origin list that has to be converted to a zero origin slice index
	ls[i-1].CompletedAt = time.Now()

	return nil
}

func (l *ListType) Delete(i int) error {
	ls := *l
	if i <= 0 || i > len(ls) {
		return fmt.Errorf("item %d does not exist to be deleted", i)
	}

	*l = append(ls[:i-1], ls[i:]...) // have to convert from 1 origin index.  And remember that item[i-1] is not part of this slice reference

	return nil
}

func (l *ListType) SaveJSON(filename string) error {
	js, err := json.Marshal(l)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, js, 0644)
}

func (l *ListType) SaveBinary(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	encoder := gob.NewEncoder(f)
	err = encoder.Encode(*l)
	return err // I want to make sure that the write operation occurs before the close operation.
}

func (l *ListType) LoadJSON(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // I wanted the client to know if file was not found, but then the tests didn't work.
			return nil
		}
		return err
	}
	if len(file) == 0 {
		return nil
	}

	return json.Unmarshal(file, l)
}

func (l *ListType) LoadBinary(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // I wanted the client to know if the file was not found, but then the tests didn't work.
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(l)
	return err
}

func (l *ListType) List() []string {
	s := make([]string, 0, len(*l))

	for _, v := range *l {
		s = append(s, v.Task)
	}
	return s
}

func (l *ListType) About() string {
	return lastModified
}

func (l *ListType) GetString(i int) string { // this is a 1 origin coordinate system.
	var formatted string

	if i < 0 || i >= len(*l) {
		return ""
	}

	suffix := ".  "
	prefix := "  "

	t := (*l)[i] // This code uses i=0 to mean no item number was given.  So have to correct from 1 origin to 0 origin.

	createdAt := ", created " + t.CreatedAt.Format("Jan-02-2006 15:04") + ", "
	if t.Done {
		prefix = "X "
		suffix = " completed at " + t.CompletedAt.Format("Jan-02-2006 15:04") + ".  "
	}
	// Need to adjust to 1-origin task numbers
	formatted += fmt.Sprintf("%s%d: %s%s%s\n", prefix, i+1, t.Task, createdAt, suffix) // notice that the string includes the newline character.
	return formatted
}

func (l ListType) String() string { // implements the fmt.Stringer interface, which must be a value receiver.  And it returns a single string, not a slice of strings.
	var formatted string
	/*
		for i, t := range l { // loop to have all tasks in the returned string
			suffix := ".  "
			prefix := "  "

			createdAt := ", created " + t.CreatedAt.Format("Jan-02-2006 15:04") + ", "

			if t.Done {
				prefix = "X "
				//suffix = " completed at " + t.CompletedAt.Format("Jan-02-2006 15:04:05") + ".  "
				suffix = " completed at " + t.CompletedAt.Format("Jan-02-2006 15:04") + ".  "
			}
			// Need to adjust to 1-origin task numbers
			formatted += fmt.Sprintf("%s%d: %s%s%s\n", prefix, i+1, t.Task, createdAt, suffix)
		}

	*/
	for i := range l {
		formatted += l.GetString(i + 1) // adjust to 1 origin task numbers.
	}
	return formatted
}
