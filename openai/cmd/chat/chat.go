package main

import (
	"bufio"
	"fmt"
	"os"
	"src/openai"
)

/*
  31 Aug 2024 -- From Linux Magazine 270 May 2023.
   5 Sep 2024 -- Added stop or exit when the text is too short, ie, len < 5 char.  Intended to include stop, exit and any other 4-letter word.
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
		text := scanner.Text()
		if len(text) < 5 {
			break
		}
		ai.PrintResp(text)
	}
}
