// mdp.go.  Markdown Preview tool to generate a valid HTML block, wrap w/ an HTML header and footer, so it can be viewed in a browser.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"os"
	"path/filepath"
)

const lastModified = "16 Jan 2022"

/*
REVISION HISTORY
-------- -------
16 Jan 22 -- Started reading chapter 3 in "Powerful Command-Line Applications in Go" by Ricardo Gerardi
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
	flag.StringVar(&filename, "file", "", "Markdown file to preview")
	flag.Parse()

	if filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	err := run(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// main() can't be tested using the Go tools.  But run() can be, so that's why run() is here as it returns values that can be used in tests.
// Here we will use golden files to validate the output, as the results can be complex as they're entire HTML files.
// A special subdir off of mdp/ is created, called testdata.  This is ignored by the Go build tools.
func run(filename string) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData := parseContent(input)

	outName := fmt.Sprintf("%s.html", filepath.Base(filename))
	fmt.Println(outName)
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
