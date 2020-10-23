package main

/*
  From packtpub book, Mastering Concurrency in Go.
  I'm confused by the use of new() below.

*/

import (
	"fmt"
	"time"
)

type Job struct {
	i    int
	max  int
	text string
}

func outputText(j *Job) {
	for j.i < j.max {
		time.Sleep(1 * time.Millisecond)
		fmt.Println(j.text)
		j.i++
	}
}

func main() {
	hello := new(Job)
	world := new(Job)
	fmt.Printf(" hello is of type %T, world is of type %T \n", hello, world)
	// this Printf showed these to be of type *main.Job.  This helps me.

	hello.text = "hello"
	hello.i = 0
	hello.max = 3

	world.text = "world"
	world.i = 0
	world.max = 5

	go outputText(hello)
	outputText(world)

	// I now realize why the use of new() looked odd to me.  The idiom is to use a different syntax.

	hello2 := &Job{}
	fmt.Printf(" hello2 is of type %T, initialized using a more idiomatic syntax. \n", hello2)
	hello2.text = "hello2"
	hello2.i = 0
	hello2.max = 5
	outputText(hello2)

}
