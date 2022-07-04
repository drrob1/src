package main

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func showTimer(t int) { // relies on creation of the "st.flg" file upon hitting escape, as the Modula-2 version does.
	timeString := strconv.Itoa(t)
	cmd := exec.Command("showtimer", timeString)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Run() // run will wait for the showtimer.exe file to finish before returning
}

func showTimerStr1(t int) string { // relies on behavior of gofshowtimer.go, which returns a string "escaped" when the <esc> is hit.
	timeString := strconv.Itoa(t)
	var strBuilder strings.Builder // this would also work to be defined as a pointer and then there is no adressOf operator on cmd.Stdoug
	cmd := exec.Command("showtimer", timeString)
	cmd.Stdin = nil
	cmd.Stdout = &strBuilder // if the strBuilder is defined as a pointer, then this line would not have the '&' operator.
	cmd.Run()                // run will wait for the showtimer.exe file to finish before returning
	inputStr1 := strBuilder.String()
	return inputStr1
}

func showTimerStr2(t int) string { // relies on behavior of gofshowtimer.go, which returns a string "escaped" when the <esc> is hit.
	timeString := strconv.Itoa(t)
	var buf *bytes.Buffer
	cmd := exec.Command("showtimer", timeString)
	cmd.Stdin = nil
	cmd.Stdout = buf
	cmd.Run()
	inputStr2 := buf.String()
	return inputStr2
}
