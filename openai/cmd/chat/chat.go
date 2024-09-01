package main

import (
	"bufio"
	"fmt"
	"os"
	"src/openai"
)

/*
  31 Aug 2024 -- From Linux Magazine 270 May 2023.
*/

func main() {
	ai := openai.NewAI()
	ai.Init()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf(" Ask: ")
		if !scanner.Scan() {
			break
		}
		ai.PrintResp(scanner.Text())
	}
}
