package main   // closing channels, to make stuff happen or not.  But now will need a stopCh

import (
        "fmt"
	"time"
	"math/rand"
       );

func reader(intCh chan int) { 
	t := time.NewTimer(10*time.Second);

	for {
	  select {
	    case i := <- intCh:    // when this channel is set to nil, it does not error out.  
				// But this case will never be satified after the channel is nil.
		fmt.Println(i);
	    case <- t.C:
		intCh = nil;   // nil used on a receiving channel
	  }
	}

}

func writer(intCh chan int) {
	stoptime := time.NewTimer(2*time.Second);
	resumetime := time.NewTimer(4*time.Second);

	savedCh := intCh;

	for {
	  select {
	    case  intCh <- rand.Intn(42):
	    case <- stoptime.C:
		intCh = nil;  // stop sending random numbers down channel when it is set to nil.
	    case <- resumetime.C:
		intCh = savedCh;
	  }
	}
}


func main() {
	intCh := make(chan int);

	go reader(intCh);
	go writer(intCh);


	time.Sleep(30*time.Second);

	fmt.Println();
}
