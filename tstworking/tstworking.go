package main

import (
	"fmt"
	"os"
)

func main() {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" Getwd failed with ERROR: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf(" Current working directory: %s\n", workingDir)

}
