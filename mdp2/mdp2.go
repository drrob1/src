// mdp2.go.  Markdown Preview tool to generate a valid HTML block, wrap w/ an HTML header and footer, so it can be viewed in a browser.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const lastModified = "17 Jan 2022"

/*
REVISION HISTORY
-------- -------
16 Jan 22 -- Started reading chapter 3 in "Powerful Command-Line Applications in Go" by Ricardo Gerardi
17 Jan 22 -- Adding use of a temporary file for outfile file.
             Now called mdp2.go, and will use interfaces to automate tests, which is the next section in the book
             Adding an auto-preview feature to this tool.
*/

const header = `<!DOCTYPE html>
<html>
      <head>
      <meta http-equiv="content-type" content="text/html; charset=utf-8">
      <title>Markdown Preview Tool</title>
      </head>
  <body>
`
const footer = `
    </body>
</html>`

func main() {
	fmt.Printf("mdp, a Markdown Previewer tool, last modified %s \n", lastModified)
	var filename string
	flag.StringVar(&filename, "f", "", "Markdown file to preview")
	skipPreview := flag.Bool("skip", false, "Skip the preview step.")
	flag.Parse()

	if filename == "" && flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	} else if flag.NArg() > 0 {
		filename = flag.Arg(0)
	}

	err := run(filename, os.Stdout, *skipPreview)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// main() can't be tested using the Go tools, as main() doesn't return anything.  But run() can be, so that's why run() is here as it returns values that can be used in tests.
// And using the run() function allows the use of a defer statement to clean up our resources.  main() relies on os.Exit() which exits immediately and does not execute any defer statements.
// Here we will use golden files to validate the output, as the results can be complex as they're entire HTML files.
// A special subdir off of mdp/ is created, called testdata.  This is ignored by the Go build tools.
func run(filename string, out io.Writer, skipPreview bool) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData := parseContent(input)

	//outName := fmt.Sprintf("%s.html", filepath.Base(filename)) Temp file allows multiple runs of same input file, likely w/ small changes.
	temp, err := os.CreateTemp("", "mdp*.html") // This sends the file into /tmp
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil { // close it now as it's not being written to yet, and make sure there are no errors upon closing it.
		return err
	}

	outName := temp.Name()
	fmt.Fprintln(out, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	if skipPreview {
		return nil
	}

	defer os.Remove(outName) // only delete this after it is to be previewed, if it's previewed.  But this is a race condition that will be fixed in the preview function.
	return preview(outName)
}

func parseContent(input []byte) []byte {
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	var buffer bytes.Buffer

	buffer.WriteString(header)
	buffer.Write(body)
	buffer.WriteString(footer)

	return buffer.Bytes()
}

func saveHTML(outFname string, data []byte) error {
	return os.WriteFile(outFname, data, 0644)
}

func preview(fname string) error {
	var cName string
	var cParams []string

	// define the exec based on OS
	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "tcc.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"

	default:
		return fmt.Errorf("OS not supported")
	}

	cParams = append(cParams, fname)

	cPath, err := exec.LookPath(cName) // Find the executable file in the PATH
	if err != nil {
		return err
	}

	err = exec.Command(cPath, cParams...).Run()

	// give the browser time to open the file before attempting to delete it.  Better solutions use a signal or a REST api to create a small web server that serves the file directly to the browser.
	time.Sleep(2 * time.Second)
	return err
}

/*
A bash script to preview a markdown file everytime it changes.  If going to use it, remember to make it executable.

#! /bin/bash
FHASH=$(md5sum $1)
while true; do
  NHASH= $(md5sum $1)
  if [ "$NHASH" != "$FHASH" ]; then
    ./mdp $1
    FHASH=$NHASH
  fi
  sleep 5  # means 5 seconds
done
*/
