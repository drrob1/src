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
)

const lastModified = "17 Jan 2022"

/*
REVISION HISTORY
-------- -------
16 Jan 22 -- Started reading chapter 3 in "Powerful Command-Line Applications in Go" by Ricardo Gerardi
17 Jan 22 -- Adding use of a temporary file for outfile file.
             Now called mdp2.go, and will use interfaces to automate tests, which is the next section in the book
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
	flag.Parse()

	if filename == "" && flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	} else if flag.NArg() > 0 {
		filename = flag.Arg(0)
	}

	err := run(filename, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// main() can't be tested using the Go tools, as main() doesn't return anything.  But run() can be, so that's why run() is here as it returns values that can be used in tests.
// Here we will use golden files to validate the output, as the results can be complex as they're entire HTML files.
// A special subdir off of mdp/ is created, called testdata.  This is ignored by the Go build tools.
func run(filename string, out io.Writer) error {
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

	return saveHTML(outName, htmlData)
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
