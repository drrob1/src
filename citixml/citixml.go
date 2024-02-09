// citixml.go
package main

import (
	"encoding/xml"
	"fmt"
	"src/getcommandline"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var t, token xml.Token
var err error

const KB = 1024
const MB = KB * KB
const qfxext = ".qfx"
const xmlext = ".xml"

type citixmlheadertype struct {
	//	DTSERVER string `xml:"DTSERVER,chardata,OFX>SIGNONMSGSRSV1>SONRS" `
	DTSERVER string `xml:"DTSERVER,OFX>SIGNONMSGSRSV1>SONRS" `
	//	LANGUAGE string `xml:"LANGUAGE,chardata,OFX>SIGNONMSGSRSV1>SONRS" `
	LANGUAGE string `xml:"LANGUAGE,OFX>SIGNONMSGSRSV1>SONRS" `
	ORG      string `xml:"ORG,OFX>SIGNONMSGSRSVR1>SONRS>FI" `
	FID      int    `xml:"FID,OFX>SIGNONMSGSRSVR1>SONRS>FI" `
	CURDEF   string `xml:"CURDEF,OFX>BANKMSGSRSV1>STMTTRNRS"`
	BANKID   string `xml:"BANKID,OFX>BANKMSGSRSV1>STMTTRNRS>BANKACCTFROM"`
	ACCTID   string `xml:"ACCTID,OFX>BANKMSGSRSV1>STMTTRNRS>BANKACCTFROM"`
	ACCTTYPE string `xml:"ACCTTYPE,OFX>BANKMSGSRSV1>STMTTRNRS>BANKACCTFROM"`
	DTSTART  string `xml:"DTSTART,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST"`
	DTEND    string `xml:"DTEND,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST"`
}

type cititransxmltype struct {
	TRNTYPE  string `xml:"TRNTYPE,chardata,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST>STMTTRN"`
	DTPOSTED string `xml:"TRNTYPE,chardata,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST>STMTTRN"`
	TRNAMT   int    `xml:"TRNTYPE,chardata,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST>STMTTRN"`
	FITID    string `xml:"TRNTYPE,chardata,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST>STMTTRN"`
	CHECKNUM int    `xml:"TRNTYPE,chardata,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST>STMTTRN"`
	NAME     string `xml:"TRNTYPE,chardata,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST>STMTTRN"`
	MEMO     string `xml:"TRNTYPE,chardata,OFX>BANKMSGSRSV1>STMTTRNRS>BANKTRANLIST>STMTTRN"`
}

func main() {

	var e error
	var citiheader citixmlheadertype
	var cititrans cititransxmltype

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

	filebyteslice := make([]byte, 0, MB) // 1 MB as initial capacity.
	filebyteslice, e = ioutil.ReadFile(InFilename)
	if e != nil {
		fmt.Println(" Error from ReadFile is ", e)
		os.Exit(1)
	}

	e = xml.Unmarshal(filebyteslice, &citiheader) // this requires a []byte
	if e != nil {
		fmt.Println(" Unmarshalling the header error", e)
		os.Exit(1)
	}
	fmt.Printf(" header %#v \n", citiheader)

	cititransactionslice := make([]cititransxmltype, 0, 500)

	for {
		e = xml.Unmarshal(filebyteslice, &cititrans) // this requires a []byte
		if e == io.EOF {
			fmt.Println("Reached EOF")
			break
		}
		if e != nil {
			fmt.Println(" Unmarshalling a transaction got error ", e)
			break
		}

		cititransactionslice = append(cititransactionslice, cititrans)
		fmt.Printf(" transaction %#v \n", cititrans)
	}
	fmt.Println(" Number of transactions ", len(cititransactionslice))
} // main

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
