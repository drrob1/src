/*
  17 Sep 23 -- Singly linked list for the live project.
*/

package main

import "fmt"

type Cell struct {
	data string
	next *Cell
}

type linkedList struct {
	sentinel *Cell
}

func main() {
	aCell := Cell{data: "Apple", next: nil}
	bCell := Cell{data: "Banana"}
	aCell.next = &bCell
	top := &aCell

	// Now to add a sentinel.  The purpose of a sentinel is to make it easy to add an item to the beginning of a linked list.  The sentinel itself never contains data,
	// just is a pointer to the next element.

	sentinel := Cell{next: top}
	top = &sentinel

	var counter int
	for cel := top; cel != nil; cel = cel.next {
		counter++
		fmt.Printf(" Cell.Data[%d]: %q  Next: %p\n", counter, cel.data, cel.next)
	}
}
