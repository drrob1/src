package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

/*
23 Apr 2022 -- From chapter 6 of the book "Powerful Command Line Applications in Go" by Ricardo Gerardi.  CI means continuous improvement.
                 That is, it's designed to recompile a program from git code automatically.  This is what continuous improvement means.

The built-in type error is an interface that defines a single method w/ the signature Error() string.  Any type that implements this will satisfy the interface.

Now the run routine will be refactored to make it easier to configure, so we'll add a custom type, step, that represents a pipeline step and associate the method
execute() to it.  And we'll add a constructor function called newStep() to create a new step.  When we need to add a new step to the build pipeline, we
instantiate the step type w/ the appropriate values.
*/

const lastModified = "Apr 23, 2022"

var ErrValidation = errors.New("validation failed")

type stepErr struct {
	step  string
	msg   string
	cause error
}

func (se *stepErr) Error() string {
	return fmt.Sprintf("Step: %q: %s: Cause: %v", se.step, se.msg, se.cause)
}

func (se *stepErr) Is(target error) bool {
	t, ok := target.(*stepErr)
	if !ok {
		return false
	}
	return t.step == se.step
}

func (se *stepErr) Unwrap() error { // attempt to unwrap the error to see if an underlying error matches the target.
	return se.cause
}

type step struct {
	name    string   // step name
	exe     string   // exe name of the external tool we need to execute
	args    []string // arguments for the executable
	message string   // output message in case of success
	proj    string   // target project on which to execute the task
}

func newStep(name, exe, message, proj string, args []string) step {
	return step{
		name:    name,
		exe:     exe,
		message: message,
		args:    args,
		proj:    proj,
	}
}

func (stp step) execute() (string, error) {
	cmd := exec.Command(stp.exe, stp.args...)
	cmd.Dir = stp.proj
	if err := cmd.Run(); err != nil {
		se := stepErr{
			step:  stp.name,
			msg:   "failed to execute",
			cause: err,
		}
		return "", &se // pointer semantics, because the Error() method is defined as having a pointer receiver.
	}
	return stp.message, nil
}

func run(proj string, out io.Writer) error {
	if proj == "" {
		//return fmt.Errorf("project directory is required")
		return fmt.Errorf("project directory is required: %w", ErrValidation)
	}

	// arguments for the go command, go build in this case.  The dot is to represent the current directory.  The go build is to verify the program's correctness to compile, rather
	// than the creation of an exe binary file.  Go build does not create a file when building multiple packages at the same time.  The other package to be built will be the errors
	// package from the standard library.
	// args := []string{"build", ".", "errors"}  Now supplanted by refactoring to be more general.
	//cmd := exec.Command("go", args...)
	//cmd.Dir = proj
	//if err := cmd.Run(); err != nil {
	//	se := stepErr{step: "go build", msg: "go build failed", cause: err} // this instantiates and returns a new stepErr w/ custom content.
	//	return &se  // makes it clear that this is pointer semantics.
	//}
	//_, err := fmt.Fprintln(out, "Go build succeeded")

	pipeline := make([]step, 1) // start w/ one step element, but more will be added soon.
	pipeline[0] = newStep(
		"go build",
		"go",
		"Go Build: SUCCESS",
		proj,
		[]string{"build", ".", "errors"},
	)

	for _, stp := range pipeline {
		msg, err := stp.execute()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	fmt.Printf(" Go CI continuous improvement, last modified %s, compiled %s\n", lastModified, runtime.Version())
	proj := flag.String("p", "", "Project directory")
	flag.Parse()

	if err := run(*proj, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
