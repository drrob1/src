// Listing 12.17 -- xml.go from one of the texts I bought.
package main

import (
	"encoding/xml"
	"fmt"
	"getcommandline"
	"os"
	"path/filepath"
	"strings"
)

var t, token xml.Token
var err error

const qfxext = ".qfx"
const xmlext = ".xml"

func main() {

	var e error

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: xmltest <hashFileName.ext> where .ext = [.qfx|.xml]")
		os.Exit(1)
	}
	inbuf := getcommandline.GetCommandLineString()
	BaseFilename := filepath.Clean(inbuf)
	InFilename := ""
	InFileExists := false

	if strings.Contains(BaseFilename, ".") { // there is an extension here
		InFilename = BaseFilename
		_, err := os.Stat(InFilename)
		if err == nil {
			InFileExists = true
		}
	} else {
		InFilename = BaseFilename + qfxext
		_, err := os.Stat(InFilename)
		if err == nil {
			InFileExists = true
		} else {
			InFilename = BaseFilename + xmlext
			_, err := os.Stat(InFilename)
			if err == nil {
				InFileExists = true
			}
		}

	}

	if !InFileExists {
		fmt.Println(" File ", BaseFilename, BaseFilename+qfxext, BaseFilename+xmlext, " or ", InFilename, " do not exist.  Exiting.")
		os.Exit(1)
	}

	InputFile, err := os.Open(InFilename)
	if err != nil {
		fmt.Println(" Error while opening ", InFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer InputFile.Close()

	p := xml.NewDecoder(InputFile)

	for t, e = p.Token(); e == nil; t, e = p.Token() {
		switch token := t.(type) {
		case xml.StartElement:
			name := token.Name.Local
			fmt.Printf("Token name: %s\n", name)
			for _, attr := range token.Attr {
				attrName := attr.Name.Local
				attrValue := attr.Value
				fmt.Printf("An attribute is: %s %s\n", attrName,
					attrValue)
				// ..
			}
		case xml.EndElement:
			fmt.Println("End of token")
		case xml.CharData:
			content := string([]byte(token))
			fmt.Printf("This is the content: %v\n", content)
			// ...
		default:
			// ...
		}
	}
	fmt.Println("Error code is ", e)
}

/* Output:
Token name: Person
Token name: FirstName
This is the content: Laura
End of token
Token name: LastName
This is the content: Lynn
End of token
End of token

The package defines a number of types for XML-tags: StartElement, Chardata (this is the actual text between the start- and end-tag), EndElement, Comment, Directive or ProcInst.
It also defines a struct Parser: the method NewParser takes an io.Reader (in this case a strings.NewReader) and produces an object of type Parser.  This has a method Token() that returns the next XML
token in the input stream.  At the end of the input stream, Token() returns nil (io.EOF).
The XML-text is walked through in a for-loop which ends when the Token() method returns an error at end of file because there is no token left anymore to parse.  Through a type-switch further processing
can be defined according to the current kind of XML-tag.  Content in Chardata is just a [ ]bytes, make it readable with a string conversion.
*/
