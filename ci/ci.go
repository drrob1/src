package main

import (
	"bytes"
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

28 Apr 2022 -- Handling output from external programs (p 179).  The version of the "execute()" method of the step type does not handle the program output.
               We will create another type called exceptionStep that extends the step type and implements another version of the "execute()" method to handle pgm output.
               This results in less complex functions that are easier to maintain, than extending the first execute() method.
                   We'll also introduce a new interface called executer that expects a single execute() method that returns a string and an error.  This interface will
               be used in the pipeline definition so that any type that satisfies the interface can be added to the pipeline.
               The gofmt tool will be used to validate the go file.  Its default behavior is to print the properly formatted version to stdout.  We will use the -l option (-el)
               which will return the name of the files that did not match the correct formatting.
*/

const lastModified = "May 1, 2022"

var ErrValidation = errors.New("validation failed")

type executer interface {
	execute() (string, error)
}

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

type exceptionStep struct { // from p 180
	step // by embedding this type, all the fields and methods of this type are available to this new type.  So can have anew version of the "execute()" method.
}

func newExceptionStep(name, exe, message, proj string, args []string) exceptionStep {
	stp := exceptionStep{}
	stp.step = newStep(name, exe, message, proj, args)

	return stp
}

func (es exceptionStep) execute() (string, error) { // extends step.execute()
	cmd := exec.Command(es.exe, es.args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = es.proj
	if err := cmd.Run(); err != nil {
		se := stepErr{ // This is the style taught by Bill Kennedy, not the style used in this book.
			step:  es.name,
			msg:   "failed to execute",
			cause: err,
		}
		return "", &se // make it clear that we're returning a pointer.
	}

	if out.Len() > 0 { // Command has run and finished.  If this buffer has content then a file errored out.
		se := stepErr{
			step:  es.name,
			msg:   fmt.Sprintf("invalid format: %s", out.String()),
			cause: nil,
		}
		return "", &se
	}

	return es.message, nil
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

	pipeline := make([]executer, 3) // started w/ one step element, the 2, and now 3.  First []step, now what is there
	pipeline[0] = newStep(
		"go build",
		"go",
		"Go Build: SUCCESS",
		proj,
		[]string{"build", ".", "errors"},
	)
	pipeline[1] = newStep(
		"go test",
		"go",
		"Go Test: SUCCESS",
		proj,
		[]string{"test", "-v"},
	)
	pipeline[2] = newExceptionStep(
		"go fmt",
		"gofmt",
		"Gofmt: SUCCESS",
		proj,
		[]string{"-l", "."},
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

	var proj string
	flag.StringVar(&proj, "p", "", "Project directory")
	flag.Parse()

	if flag.NArg() > 0 {
		proj = flag.Arg(0)
	}

	if err := run(proj, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
