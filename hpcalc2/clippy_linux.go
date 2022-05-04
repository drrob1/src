package hpcalc2

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func toClip(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	rdr := strings.NewReader(s)
	cmd := exec.Command("xclip")
	cmd.Stdin = rdr
	cmd.Stdout = os.Stdout
	cmd.Run()
	return fmt.Sprintf(" Sent %s to xclip.", s)
}

func fromClip() (float64, string, error) {
	w := bytes.NewBuffer([]byte{})
	cmd := exec.Command("xclip", "-o") // that an oh flag
	cmd.Stdout = w
	cmd.Run()
	str := w.String()
	s := strings.ReplaceAll(str, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	msg := fmt.Sprintf(" Received %s from xclip.  After removing cleaning it up it became %s.", str, s)
	f, err := strconv.ParseFloat(str, 64)
	return f, msg, err
}

/*
{{{
case 520: // TOCLIP
	R := READX()
	s := strconv.FormatFloat(R, 'g', -1, 64)
	if runtime.GOOS == "linux" {
		linuxclippy := func(s string) {
			buf := []byte(s)
			rdr := bytes.NewReader(buf)
			cmd := exec.Command("xclip")
			cmd.Stdin = rdr
			cmd.Stdout = os.Stdout
			cmd.Run()
			ss = append(ss, fmt.Sprintf(" Sent %s to xclip.", s))
		}
		linuxclippy(s)
	} else if runtime.GOOS == "windows" {
		comspec, ok := os.LookupEnv("ComSpec")
		if !ok {
			ss = append(ss, " Environment does not have ComSpec entry.  ToClip unsuccessful.")
			break outerloop
		}
		winclippy := func(s string) {
			//cmd := exec.Command("c:/Program Files/JPSoft/tcmd22/tcc.exe", "-C", "echo", s, ">clip:")
			cmd := exec.Command(comspec, "-C", "echo", s, ">clip:")
			cmd.Stdout = os.Stdout
			cmd.Run()
			ss = append(ss, fmt.Sprintf(" Sent %s to %s.", s, comspec))
		}
		winclippy(s)
	}

case 530: // FROMCLIP
	PushMatrixStacks()
	LastX = Stack[X]
	w := bytes.NewBuffer([]byte{}) // From "Go Standard Library Cookbook" as referenced above.
	if runtime.GOOS == "linux" {
		cmdfromclip := exec.Command("xclip", "-o")
		cmdfromclip.Stdout = w
		cmdfromclip.Run()
		str := w.String()
		s := fmt.Sprintf(" Received %s from xclip.", str)
		str = strings.ReplaceAll(str, "\n", "")
		str = strings.ReplaceAll(str, "\r", "")
		str = strings.ReplaceAll(str, ",", "")
		str = strings.ReplaceAll(str, " ", "")
		s = s + fmt.Sprintf("  After removing all commas and spaces it becomes %s.", str)
		ss = append(ss, s)
		R, err := strconv.ParseFloat(str, 64)
		if err != nil {
			ss = append(ss, fmt.Sprintln(" fromclip on linux conversion returned error", err, ".  Value ignored."))
		} else {
			PUSHX(R)
		}
	} else if runtime.GOOS == "windows" {
		comspec, ok := os.LookupEnv("ComSpec")
		if !ok {
			ss = append(ss, " Environment does not have ComSpec entry.  FromClip unsuccessful.")
			break outerloop
		}

		cmdfromclip := exec.Command(comspec, "-C", "echo", "%@clip[0]")
		cmdfromclip.Stdout = w
		cmdfromclip.Run()
		lines := w.String()
		s := fmt.Sprint(" Received ", lines, "from ", comspec)
		linessplit := strings.Split(lines, "\n")
		str := strings.ReplaceAll(linessplit[1], "\"", "")
		str = strings.ReplaceAll(str, "\n", "")
		str = strings.ReplaceAll(str, "\r", "")
		str = strings.ReplaceAll(str, ",", "")
		str = strings.ReplaceAll(str, " ", "")
		s = s + fmt.Sprintln(", after post processing the string becomes", str)
		ss = append(ss, s)
		R, err := strconv.ParseFloat(str, 64)
		if err != nil {
			ss = append(ss, fmt.Sprintln(" fromclip", err, ".  Value ignored."))
		} else {
			PUSHX(R)
		}
	}
}}}
*/
