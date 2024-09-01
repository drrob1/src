package main

import (
	"bufio"
	"os"
	"src/openai"
)

/*
  31 Aug 2024 -- From Linux Magazine 270 May 2023
*/

func main() {
	ai := openai.NewAI()
	ai.Init()
	scanner := bufio.NewScanner(os.Stdin)

	var text string

	for scanner.Scan() {
		text += scanner.Text()
	}
	ai.PrintResp("Translate to Spanish:\n" + text)
}
