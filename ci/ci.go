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

func run(proj string, out io.Writer) error {
	if proj == "" {
		//return fmt.Errorf("project directory is required")
		return fmt.Errorf("project directory is required: %w", ErrValidation)
	}

	// arguments for the go command, go build in this case.  The dot is to represent the current directory.  The go build is to verify the program's correctness to compile, rather
	// than the creation of an exe binary file.  Go build does not create a file when building multiple packages at the same time.  The other package to be built will be the errors
	// package from the standard library.
	args := []string{"build", ".", "errors"}

	cmd := exec.Command("go", args...)
	cmd.Dir = proj
	if err := cmd.Run(); err != nil {
		return &stepErr{step: "go build", msg: "go build failed", cause: err} // this instantiates and returns a new stepErr w/ custom content.
		//return fmt.Errorf(" go build failed: %s", err)   Initial version of this code, now expanded to use custom error types.
	}
	_, err := fmt.Fprintln(out, "Go build succeeded")
	return err
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
