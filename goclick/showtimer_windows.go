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

func gShowTimer1(execStr string, t int) string {
	timeString := strconv.Itoa(t)
	var strBuilder strings.Builder // this would also work to be defined as a pointer and then there is no adressOf operator on cmd.Stdoug
	//cmd := exec.Command("gofshowtimer", "-t", timeString)
	cmd := exec.Command(execStr, "-t", timeString)
	cmd.Stdin = nil
	cmd.Stdout = &strBuilder // if the strBuilder is defined as a pointer, then this line would not have the '&' operator.
	cmd.Run()                // run will wait for the showtimer.exe file to finish before returning
	inputStr1 := strBuilder.String()
	return inputStr1
}

func gShowTimer2(execStr string, t int) string {
	// bytes.buffer needs to be initialized, while the strings.Builder does not.  Advantage goes to strings.Builder.
	// And the initialization make call needs to be len=0, cap= whatever.
	// It didn't work when I did not make the initial len=0.  Or I can just use the empty type constant.
	timeString := strconv.Itoa(t)
	buf := bytes.NewBuffer(make([]byte, 0, 100)) // I prefer this form of the initialization.
	//buf := bytes.NewBuffer([]byte{})            // This does work.
	//buf := bytes.NewBuffer(make([]byte, 100))   // This did not work.
	//cmd := exec.Command("gofshowtimer", "-t", timeString)
	cmd := exec.Command(execStr, "-t", timeString)
	cmd.Stdin = nil
	cmd.Stdout = buf
	cmd.Run()
	inputStr2 := buf.String()
	return inputStr2
}
