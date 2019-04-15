package main

import (
	"fmt"
	"math/rand"
)
import format "fmt"

func makerandoms(c chan int) { // ns is a dynamically allocated strings slice to be sent down a channel

	for { // infinite loop
		c <- rand.Intn(1000)
	}
}

func makeID(idChan chan int) {
	var id int

	for {
		idChan <- id
		id++
	}
}

func main() {
	var limit int
	randoms := make(chan int)

	go makerandoms(randoms)

	for randint := range randoms { // range over a channel waits until the channel is closed.
		fmt.Printf("%d ", randint)
		limit++
		if limit > 200 { // this would go on infinitely if I did not do this.
			break
		}
	}

	format.Println()

	//	go makerandoms(randoms);  don't need to open this again because I never closed the randoms
	//	channel
	n := <-randoms
	fmt.Println(n)
	n = <-randoms
	fmt.Println(n)
	n = <-randoms
	fmt.Println(n)
	n = <-randoms
	fmt.Println(n)
	n = <-randoms
	fmt.Println(n)

	idChan := make(chan int)
	go makeID(idChan)
	format.Println(<-idChan)

	format.Println(<-idChan)
	format.Println(<-idChan)
	format.Println(<-idChan)
	format.Println(<-idChan)
	format.Println(<-idChan)
}
