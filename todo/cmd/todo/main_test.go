package main_test

import (
	"fmt"
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
		cmd := exec.Command(cmdPath, "-task", task)
		err = cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Run("AddNewTask", addNewTaskTestFunc)

	listTasksTestFunc := func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		//expected := " Not done: " + task + "\n\n"
		expected := "  1: " + task + ".  \n\n"
		if expected != string(out) {
			t.Errorf(" Expected %q, got %q instead\n", expected, string(out))
		}

	}
	t.Run("ListTasks", listTasksTestFunc)
}
