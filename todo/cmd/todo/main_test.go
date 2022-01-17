package main_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var execName = "todo"

var fileName = "test-todo"
var binFilename = "test-todo"

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		execName += ".exe"
	}

	os.Setenv("TODO_FILENAME", fileName)
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("in TestMain. UserHomeDir error is", err)
	}
	fileName = filepath.Join(userHomeDir, fileName) + ".json"
	binFilename = filepath.Join(userHomeDir, binFilename) + ".gob"

	// We are compiling the tool and running it in the test cases because this code is in main() which can't be tested
	// because it doesn't return anything that can be tested using the Go tools.
	build := exec.Command("go", "build", "-o", execName)
	err = build.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Cannot build tool %s: %s \n", execName, err)
		os.Exit(1)
	}
	fmt.Println(" Running tests ....")
	result := m.Run()
	fmt.Println(" Cleaning up ....")
	err = os.Remove(execName)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from Remove(%s) is %v\n", execName, err)
	}
	err = os.Remove(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from Remove(%s) is %v\n", fileName, err)
	}

	err = os.Remove(binFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from Remove(%s) is %v\n", binFilename, err)
	}

	os.Exit(result)
}

func TestTodoCLI(t *testing.T) {
	task := "test task number 1"
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmdPath := filepath.Join(dir, execName)

	//binFilename = filepath.Join(userHomeDir, binFilename)

	addNewTaskTestFunc := func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add", task)
		err = cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Run("AddNewTaskFromArgs", addNewTaskTestFunc)

	task2 := "test task number 2"
	fromStdinFunc := func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		cmdStdIn, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}

		io.WriteString(cmdStdIn, task2)
		cmdStdIn.Close()

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("AddNewTaskFromSTDIN", fromStdinFunc)

	listTasksTestFunc := func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		//expected := " Not done: " + task + "\n\n"
		expected := "  1: " + task + ".  \n  2: " + task2 + ".  \n\n"
		if expected != string(out) {
			t.Errorf(" Expected %q, got %q instead\n", expected, string(out))
		}

	}
	t.Run("ListTasks", listTasksTestFunc)
}
