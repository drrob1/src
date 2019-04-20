package main // closing channels, to make stuff happen or not.  But now will need a stopCh

import (
	"fmt"
	"time"
)

func printout(msg string, stopCh chan bool) {
	for { // infinite loop will keep doing default until it receives anything
		// on the stopCh, even a close message.
		select {
		case <-stopCh:
			return
		default:
			fmt.Println(msg)
		}
	}

}

func main() {
	stopCh := make(chan bool)

	for i := 0; i < 10; i++ { // these will all be waiting for their message to do start work.
		go printout(fmt.Sprintf("printout: %d", i), stopCh)

	}

	time.Sleep(2 * time.Second)
	close(stopCh) // this will make all the waiting printout rtns get a message to close,
	// and it will print out its pending message, which was waiting for anything
	// to be received on the channel called stopCh.

	time.Sleep(2 * time.Second) // wait for all the message to come out and be seen before exiting.

	fmt.Println()
}
