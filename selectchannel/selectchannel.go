package main

import (
	"fmt"
)

func emit(wordchannel chan string, done chan bool) {
	defer close(wordchannel)
	words := []string{"The", "quick", "brown", "fox"}
	var i int // same as if I had done i := 0; which is what the video code does

	for { // infinitely
		select {
		case wordchannel <- words[i]: // keep sending words in rotation on this channel
			i++
			if i == len(words) {
				i = 0
			}
		case <-done: // if receive a message on this channel, return
			done <- true
			close(done)
			return
		}

	}
}

func main() {

	wordch := make(chan string)
	donech := make(chan bool)

	go emit(wordch, donech)
	for i := 0; i < 100; i++ {
		fmt.Printf("%s ", <-wordch)
	}

	donech <- true

	<-donech // this will just wait until it gets a message from this channel, and then discard it.

	fmt.Println()
}
