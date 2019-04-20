package main

import (
	"fmt"
	"time"
)

func emit(chanChannel chan chan string, done chan bool) {
	wordChannel := make(chan string)
	chanChannel <- wordChannel
	defer close(wordChannel)
	defer close(done)
	words := []string{"The", "quick", "brown", "fox"}
	var i int // same as if I had done i := 0; which is what the video code does

	t := time.NewTimer(1 * time.Second)

	for { // infinitely
		select {
		case wordChannel <- words[i]: // keep sending words in rotation on this channel
			i++
			if i == len(words) {
				i = 0
			}
		case <-done: // if receive a message on this channel, return
			done <- true
			//			close(done); seems to not be needed when we send a message back on this channel
			return
		case <-t.C:
			//			done <- true; // This line deadlocked when the timer went off.  ???
			return
		}

	}
}

func main() {

	channelCh := make(chan chan string)
	doneCh := make(chan bool)

	go emit(channelCh, doneCh)

	wordCh := <-channelCh
	for word := range wordCh {
		fmt.Printf("%s ", word)
	}

	fmt.Println()
}
