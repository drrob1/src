package todocore

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

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
		return fmt.Errorf("Item %d does not exist", i)
	}

	ls[i-1].Done = true // remember i is a 1 origin list that has to be converted to a zero origin slice index
	ls[i-1].CompletedAt = time.Now()

	return nil
}

func (l *ListType) Delete(i int) error {
	ls := *l
	if i <= 0 || i > len(ls) {
		return fmt.Errorf("Item %d does not exist to be deleted.", i)
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
		if errors.Is(err, os.ErrNotExist) {
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
		if errors.Is(err, os.ErrNotExist) {
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
