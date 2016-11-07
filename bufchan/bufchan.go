// buffered channels section.  Can send on a buffered channel without a receiver being ready.
// Can receive on a buffered channel without a sender being ready to send at that moment.
// Up to the defined capacity of the buffer.
// He suggests using unbuffered channels primarily because it makes program logic and debugging much
// easier.

package main  

import (
        "fmt"
	"time"
	"math/rand"
	"sync/atomic"
       );

var (
	running int64 = 0;
)

func work() {
	atomic.AddInt64(&running,1);
	fmt.Printf("[%d",running);
	time.Sleep(time.Duration(rand.Intn(2))*time.Second);
	atomic.AddInt64(&running,-1);
	fmt.Printf("]");
}

func worker(semaphore chan bool) {
	<- semaphore   // only do something when can receive from the semaphore channel.
	work();
	semaphore <- true; // when done, put back into the channel
}



func main() {
//	intCh := make(chan int);  // unbuffered
	semaphore := make(chan bool, 10); // buffered, this channel can store up to 10 int's as defined here.

	for i := 0; i < 1000; i++ { // will start 1000, but only 10 can receive at a time from the buffer
	  go worker(semaphore);
	}

	for i := 0; i < cap(semaphore); i++ {
	  semaphore <- true;  // this won't block because of the buffer that can fill and accept these.
	}

	time.Sleep(30*time.Second);

	fmt.Println();
}
