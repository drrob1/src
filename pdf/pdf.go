package main

import (
	"bufio"
	"bytes"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/ledongthuc/pdf"
	"os"
	"strings"
)

/*
  23 Jan 25 -- First started working on this
*/

const lastModified = "25 Jan 2025"
const debugFilename = "debugPDF.txt"

func main() {

	if len(os.Args) < 2 { // first arg is always the binary name being executed
		ctfmt.Printf(ct.Red, true, "\n Need name of pdf file on command line\n")
		return
	}
	filename := os.Args[1]
	s2 := fmt.Sprintf(" pdf reading program for %s, last modified %s", filename, lastModified)
	fmt.Println(s2)
	//fmt.Printf(" pdf reading program for %s, last modified = %s.\n", filename, lastModified)

	//pdf.DebugOn = true
	//content, err := readPdf(filename)
	//if err != nil {
	//	ctfmt.Printf(ct.Red, true, err.Error())
	//	return
	//}
	//
	//fmt.Printf("\npdf content length: %d\n%s\n", len(content), content)
	//
	//if pause() {
	//	return
	//}
	//
	//content, err = readPdf2(filename)
	//if err != nil {
	//	ctfmt.Printf(ct.Red, true, err.Error())
	//	return
	//}
	//
	//if pause() {
	//	return
	//}

	//content, err := readPdfByRow(filename)
	//if err != nil {
	//	ctfmt.Printf(ct.Red, true, err.Error())
	//	return
	//}
	//fmt.Printf("\npdf content length: %d\n", len(content))

	content, err := readPdfByCol(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, err.Error())
		return
	}
	fmt.Printf("\npdf content length: %d\n", len(content))

}

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	// remember close file
	if err != nil {
		return "", err
	}
	defer f.Close()
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

func readPdf2(path string) (string, error) {
	f, r, err := pdf.Open(path)
	// remember close file
	if err != nil {
		return "", err
	}
	defer f.Close()
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		var lastTextStyle pdf.Text
		texts := p.Content().Text
		for _, text := range texts {
			if isSameSentence(text, lastTextStyle) {
				lastTextStyle.S = lastTextStyle.S + text.S
			} else {
				fmt.Printf("Font: %s, Font-size: %f, x: %f, y: %f, content: %s \n", lastTextStyle.Font, lastTextStyle.FontSize, lastTextStyle.X, lastTextStyle.Y, lastTextStyle.S)
				lastTextStyle = text
			}
		}
	}
	return "", nil
}

func isSameSentence(sentence, style pdf.Text) bool {
	if sentence.Font == style.Font && sentence.FontSize == style.FontSize {
		return true
	}
	return false
}

func readPdfByCol(path string) (string, error) {
	f, r, err := pdf.Open(path)

	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	totalPage := r.NumPage()
	fmt.Printf(" totalPage: %d\n", totalPage)

	debug, err := os.Create(debugFilename)
	if err != nil {
		return "", err
	}
	defer debug.Close()

	debugBuf := bufio.NewWriter(debug)
	defer debugBuf.Flush()

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	s2 := fmt.Sprintf(" pdf reading program for %s, last modified %s, timpstamp %s\n\n", path, lastModified, ExecTimeStamp)
	debugBuf.WriteString(s2)

	for pageIndex := range totalPage {
		p := r.Page(pageIndex + 1) // there is no page zero
		if p.V.IsNull() {
			continue
		}

		cols, _ := p.GetTextByColumn()
		for i, col := range cols {
			var bldr strings.Builder
			fmt.Printf(">>>> col#: %d; col.Position: %d\n", i, col.Position)
			s := fmt.Sprintf(">>>> col#: %d; col.Position: %d\n", i, col.Position)
			debugBuf.WriteString(s)
			for row, word := range col.Content {
				fmt.Println(word.S, "|")
				s1 := fmt.Sprintf("%s: row #%d in col# %d:X=%.3f:Y=%.3f:W=%.3f /// ", word.S, row, i, word.X, word.Y, word.W)
				debugBuf.WriteString(s1)
				bldr.WriteString(word.S)
			}
			s = fmt.Sprintf("\nword: %q\n", bldr.String())
			debugBuf.WriteString(s)
			debugBuf.WriteRune('\n')
		}
	}
	return "", nil
}

func readPdfByRow(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	totalPage := r.NumPage()
	fmt.Printf(" totalPage: %d\n", totalPage)

	debug, err := os.Create(debugFilename)
	if err != nil {
		return "", err
	}
	defer debug.Close()

	debugBuf := bufio.NewWriter(debug)
	defer debugBuf.Flush()

	for pageIndex := range totalPage {
		p := r.Page(pageIndex + 1) // there is no page zero
		if p.V.IsNull() {
			continue
		}

		rows, _ := p.GetTextByRow()
		for i, row := range rows {
			var bldr strings.Builder
			fmt.Printf(">>>> row#: %d; row.Position: %d\n", i, row.Position)
			s := fmt.Sprintf(">>>> row#: %d; row.Position: %d\n", i, row.Position)
			debugBuf.WriteString(s)
			for j, word := range row.Content {
				fmt.Println(word.S, "|")
				s1 := fmt.Sprintf("%s:jcol #%d:X=%f:Y=%f:W=%f||", word.S, j, word.X, word.Y, word.W)
				debugBuf.WriteString(s1)
				bldr.WriteString(word.S)
			}
			s = fmt.Sprintf("\nword: %s\n", bldr.String())
			debugBuf.WriteString(s)
			//if pause() {
			//	return "", fmt.Errorf("page %d row %d paused", pageIndex+1, i)
			//}
			debugBuf.WriteRune('\n')
		}
	}
	return "", nil
}

func pause() bool {
	fmt.Print(" Pausing the loop.  Hit <enter> to continue; 'n' or 'x' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "x") {
		return true
	}
	return false
}

func paused() {
	fmt.Printf(" Pausing ... Hit <enter> to continue\n  ")
	var ans string
	fmt.Scanln(&ans)
}
