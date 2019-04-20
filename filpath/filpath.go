package main

import (
	"fmt"
	"getcommandline"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"timlibg"
)

func main() {

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: filpath <testhashFileName.ext>")
		os.Exit(0)
	}
	inbuf := getcommandline.GetCommandLineString()

	cleanfilename := filepath.Clean(inbuf)
	basefilename := filepath.Base(cleanfilename)
	dirfilename := filepath.Dir(cleanfilename)
	extension := filepath.Ext(inbuf)
	dirsplitfilename, namesplitfilename := filepath.Split(cleanfilename)

	fmt.Printf(" cleanfilename: %s,\n basefilename: %s,\n dirfilename: %s,\n extension: %s\n  Dir from split: %s, Name from split: %s\n", cleanfilename, basefilename, dirfilename, extension, dirsplitfilename, namesplitfilename)

	lastIndex := strings.LastIndex(basefilename, ".")
	base := basefilename[:lastIndex]
	fmt.Println(" position of last dot is :", lastIndex, ".  base without extension is:", base)

	aDateString := MakeDateStr()
	fmt.Println(" Today's constructed date string is:", aDateString)

	fmt.Println()

}

func MakeDateStr() (datestr string) {

	const DateSepChar = "-"

	m, d, y := timlibg.TIME2MDY()
	timenow := timlibg.GetDateTime()

	MSTR := strconv.Itoa(m)
	DSTR := strconv.Itoa(d)
	YSTR := strconv.Itoa(y)
	Hr := strconv.Itoa(timenow.Hours)
	Min := strconv.Itoa(timenow.Minutes)
	Sec := strconv.Itoa(timenow.Seconds)

	datestr = MSTR + DateSepChar + DSTR + DateSepChar + YSTR + "_" + Hr + DateSepChar + Min + DateSepChar +
		Sec + "__" + timenow.DayOfWeekStr
	return datestr
} // MakeDateStr
