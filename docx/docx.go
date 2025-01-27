/*
  Turned out that this didn't capture all the content from the original PDF.  I noticed that it missed the MDs off from Wed and Thurs of the test file.
  I stopped developing this path.  I ended up realizing that I could use the web interface of Acrobat from work.
*/

package main

import (
	"fmt"
	"os"

	"github.com/fumiama/go-docx"
)

func main() {
	fname := os.Args[1]
	readFile, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	fmt.Println("Reading " + fname)
	fileinfo, err := readFile.Stat()
	if err != nil {
		panic(err)
	}
	size := fileinfo.Size()
	doc, err := docx.Parse(readFile, size)
	if err != nil {
		panic(err)
	}
	fmt.Printf(" File name is %s, filesize = %d, Plain text:\n", fname, fileinfo.Size())
	for _, it := range doc.Document.Body.Items {
		switch it.(type) {
		case *docx.Paragraph, *docx.Table: // printable
			fmt.Println(it)
		}
	}
}
