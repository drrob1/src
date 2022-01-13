package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binName = "todo"
var fileName = "todo.json"
var binFilename = "todo.gob"

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	build := exec.Command("go", "build", "-o", binName)
	err := build.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Cannot build tool %s: %s \n", binName, err)
		os.Exit(1)
	}
	fmt.Println(" Running tests ....")
	result := m.Run()
	fmt.Println(" Cleaning up ....")
	os.Remove(binName)
	os.Remove(fileName)
	os.Remove(binFilename)
	os.Exit(result)
}

func TestTodoCLI(t *testing.T) {
	task := "test task number 1"
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cmdPath := filepath.Join(dir, binName)

	addNewTaskTestFunc := func(t *testing.T) {
		cmd := exec.Command(cmdPath, strings.Split(task, " ")...)
		err = cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Run("AddNewTask", addNewTaskTestFunc)

	listTasksTestFunc := func(t *testing.T) {
		cmd := exec.Command(cmdPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := task + "\n"
		if expected != string(out) {
			t.Errorf(" Expected %q, got %q instead\n", expected, string(out))
		}

	}
	t.Run("ListTasks", listTasksTestFunc)
}
