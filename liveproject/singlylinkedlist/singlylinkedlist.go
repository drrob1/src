/*
  17 Sep 23 -- Singly linked list for the live project.
*/

package main

import (
	"fmt"
	"strings"
)

type Cell struct {
	data string
	next *Cell
}

type linkedList struct {
	sentinel *Cell
}

func makeLinkedList() *Cell {
	// This is a factory function to create a new linked list, init it's sentinal pointer to a new cell, and return the linked list.
	sentinel := Cell{}
	return &sentinel
}

func (me *Cell) addAfter(after *Cell) {
	// add a Cell after me
	// Eg: aCell := Cell{"apple", nil}
	//     bCell := Cell{data: "banana"}
	//     (&aCell).addAfter(&bCell)   or  aCell.addAfter(&bCell)  this uses the syntatic sugar of Go.
	after.next = me.next // so the next field of "after" now points to wherever "me" was pointing to, ie, the next element.  So this inserts "after" between "me" and the next linked cell.
	me.next = after
}

func (me *Cell) deleteAfter() *Cell { // need to return the deleted cell.  If there is no cell after "me", panic.
	deletedCell := me.next
	if deletedCell == nil {
		panic(" no cell to after me to delete.")
	}
	me.next = deletedCell.next
	return deletedCell
}

func (list *linkedList) addRange(values []string) {
	// add the strings as cells to the end of the linked list, ie, append the new strings
	// First, find the end of the linked list
	var lastCell *Cell

	cell := list.sentinel.next
	lastCell = cell // there may not be any cells after the sentinel.
	for ; cell != nil; cell = cell.next {
		lastCell = cell.next
	}

	for _, s := range values {
		anotherCell := Cell{data: s}
		lastCell.addAfter(&anotherCell)
	}
}

func (list *linkedList) toString(separator string) string {
	var sb strings.Builder

	sentinel := list.sentinel
	if sentinel.next == nil {
		return "" // the empty string is different than a nil string.
	}

	for cell := sentinel.next; cell != nil; cell = cell.next {
		sb.WriteString(cell.data)
		sb.WriteString(separator)
	}

	finalStr := sb.String()
	if finalStr == "" {
		return finalStr
	}

	if separator == "" {
		return separator
	}
	return finalStr[:len(finalStr)-1] // don't return the final separator
}

func (list *linkedList) toSlice() []string {
	var stringSlice []string

	sentinel := list.sentinel
	if sentinel.next == nil {
		return nil
	}

	for cell := sentinel.next; cell != nil; cell = cell.next {
		stringSlice = append(stringSlice, cell.data)
	}

	return stringSlice
}

func (list *linkedList) length() int {
	if list.sentinel.next == nil {
		return 0
	}

	var counter int
	for cell := list.sentinel.next; cell != nil; cell = cell.next {
		counter++
	}

	return counter
}

func (list *linkedList) isEmpty() bool {
	return list.sentinel.next == nil
}

func (list *linkedList) push(s string) {
	var lastCell *Cell

	cell := list.sentinel.next
	lastCell = cell // there may not be any cells after the sentinel.
	for ; cell != nil; cell = cell.next {
		lastCell = cell.next
	}

	anotherCell := Cell{data: s}
	lastCell.addAfter(&anotherCell)
}

func (list *linkedList) pop() string {

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
