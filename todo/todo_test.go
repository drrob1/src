package todo_test

import (
	"fmt"
	"os"
	"src/todo"
	"testing"
)

func TestAdd(t *testing.T) {
	l := todo.ListType{}

	taskName := "New Task"
	l.Add(taskName)
	if l[0].Task != taskName {
		t.Errorf(" Expected %q, got %q instead.\n", taskName, l[0].Task)
	}

	if l[0].Done {
		t.Errorf(" New task should not be marked as completed.\n")
	}

	_ = l.Complete(1)

	if !l[0].Done {
		t.Errorf(" New task should be marked as completed now.\n")
	}
}

func TestDelete(t *testing.T) {
	l := todo.ListType{}

	tasks := []string{
		"New task 1",
		"new task 2",
		"new task 3",
		"New task 4",
	}

	for _, v := range tasks {
		l.Add(v)
	}

	if l[0].Task != tasks[0] {
		t.Errorf(" Expected %q, got %q instead. \n", tasks[0], l[0].Task)
	}

	_ = l.Delete(2)

	if len(l) != 3 {
		t.Errorf(" Expected list length %d, got %d instead.\n", 3, len(l))
	}

	if l[2].Task != tasks[3] {
		t.Errorf(" Expected %q, got %q instead.\n", tasks[3], l[2].Task)
	}
}

func TestList(t *testing.T) {
	l := todo.ListType{}

	tasks := []string{
		"New task 1",
		"new task 2",
		"new task 3",
		"New task 4",
	}

	for _, v := range tasks {
		l.Add(v)
	}

	s := l.List()

	if s[3] != tasks[3] {
		t.Errorf(" Testing List.  Expected %q, got %q instead.  List item is %q.\n", tasks[3], s[3], l[3].Task)
		fmt.Printf(" tasks = %v \n, List returned = %v\n, l is %v\n", tasks, s, l)
	}
}

func TestSaveOpen(t *testing.T) {
	l1 := todo.ListType{}
	l2 := todo.ListType{}

	taskName := "New Task"

	l1.Add(taskName)
	if l1[0].Task != taskName {
		t.Errorf(" Expected %q, got %q instead.\n", taskName, l1[0].Task)
	}

	tempfile, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf(" Error creating temp file is %v. \n", err)
	}
	defer os.Remove(tempfile.Name())

	if err = l1.SaveJSON(tempfile.Name()); err != nil {
		t.Fatalf(" Error saving list to file is %v.\n", err)
	}

	if err = l2.LoadJSON(tempfile.Name()); err != nil {
		t.Fatalf(" Error opening list from file is %v.\n", err)
	}

	if len(l1) != len(l2) {
		t.Errorf(" Length of list 1 is %d which does not match length of list 2 which is %d.\n", len(l1), len(l2))
	}

	if l1[0].Task != l2[0].Task {
		t.Errorf(" Task %q from list1 does not match %q from list2.\n", l1[0].Task, l2[0].Task)
	}

	if l1[0].CreatedAt != l2[0].CreatedAt {
		t.Errorf("Using SaveJSON Created at from list1 does not match created at from list2: list1 %#v, list2 %#v\n", l1[0].CreatedAt, l2[0].CreatedAt)
		t.Errorf("And L1 %v, L2 %v\n", l1[0].CreatedAt, l2[0].CreatedAt)
	}
	/*  I wanted to see how json writes out time values.
	if err = l1.SaveJSON("MyJSON.json"); err != nil {
		t.Errorf(" Error from SaveJSON is %v\n", err)
	}

	And then I want to see what happens when I write out and read back in a value of time.Time.  They original list1 has a value called m=+0.002-ish that is not part of list2.
	So this 'm' value does not get written out and restored, so the test fails because of this 'm' factor.
	*/
}

func TestSaveBinary(t *testing.T) {
	l1 := todo.ListType{}
	l2 := todo.ListType{}

	taskName := "New Task"
	anotherTask := "Another Task"

	l1.Add(taskName)
	l1.Add(anotherTask)
	if l1[0].Task != taskName {
		t.Errorf(" Expected %q, got %q instead.\n", taskName, l1[0].Task)
	}

	tempfile, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf(" Error creating temp file is %v. \n", err)
	}
	defer os.Remove(tempfile.Name())

	if err = l1.SaveBinary(tempfile.Name()); err != nil {
		t.Fatalf(" Error saving list to file is %v.\n", err)
	}

	if err = l2.LoadBinary(tempfile.Name()); err != nil {
		t.Fatalf(" Error opening list from file is %v.\n", err)
	}

	if len(l1) != len(l2) {
		t.Errorf(" Length of list 1 is %d which does not match length of list 2 which is %d.\n", len(l1), len(l2))
	}

	if l1[0].Task != l2[0].Task {
		t.Errorf(" Task %q from list1 does not match %q from list2.\n", l1[0].Task, l2[0].Task)
	}

	if l1[0].CreatedAt != l2[0].CreatedAt {
		t.Errorf("Using SaveBinary, Created at from list1 does not match created at from list2: list1 %#v, list2 %#v\n", l1[0].CreatedAt, l2[0].CreatedAt)
		t.Errorf("And L1 %v, L2 %v\n", l1[0].CreatedAt, l2[0].CreatedAt)
	}

}
