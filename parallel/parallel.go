package main

import (
	"fmt"
	//         "os"
)
import format "fmt"

func emit(c chan string) { // words is a dynamically allocated strings slice to be sent down a channel
	words := []string{"the", "quick", "brown", "fox"}

	for _, word := range words {
		c <- word
	}
	close(c)
}

func Emit(c2 chan string) { // words is a dynamically allocated strings slice to be sent down a channel
	words := []string{"jumped", "over", "the", "lazy", "dog"}

	for _, word := range words {
		c2 <- word
	}
	close(c2)
}

func main() {
	wordchannel := make(chan string)
	word_channel := make(chan string)

	go emit(wordchannel)

	for word := range wordchannel { // range over a channel waits until the channel is closed.
		fmt.Printf("%s ", word) // word is defined only in this for loop.
	}

	format.Println()

	go Emit(word_channel)
	word := <-word_channel // receive from wordchannel a string, as wordchannel is a string channel
	fmt.Println(word)
	word = <-word_channel // receive from wordchannel a string, as wordchannel is a string channel
	fmt.Println(word)
	word = <-word_channel // receive from wordchannel a string, as wordchannel is a string channel
	fmt.Println(word)
	word = <-word_channel // receive from wordchannel a string, as wordchannel is a string channel
	fmt.Println(word)
	word = <-word_channel // receive from wordchannel a string, as wordchannel is a string channel
	fmt.Println(word)
	word, ok := <-word_channel // receive from wordchannel a string, as wordchannel is a string channel
	fmt.Printf("%s  %t \n", word, ok)
}
