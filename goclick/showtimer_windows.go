package main

import (
	"os"
	"os/exec"
	"strconv"
)

func showTimer(t int) {
	timeString := strconv.Itoa(t)
	cmd := exec.Command("showtimer", timeString)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // run will wait for the showtimer.exe file to finish before returning
}
