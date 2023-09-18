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

func makeLinkedList() linkedList {
	// This is a factory function to create a new linked list, init it's sentinel pointer to a new cell, and return the linked list.
	list := linkedList{}
	list.sentinel = &Cell{"SENTINAL", nil}
	//fmt.Printf(" in makeLinkedList.  List: %+v\n", list)
	return list
}

func (me *Cell) addAfter(after *Cell) {
	// add a Cell after me
	// Eg: aCell := Cell{"apple", nil}
	//     bCell := Cell{data: "banana"}
	//     (&aCell).addAfter(&bCell)   or  aCell.addAfter(&bCell)  this uses the syntatic sugar of Go.
	//fmt.Printf(" before assignment: me = %+v, after = %+v\n", me, after)
	after.next = me.next // so the next field of "after" now points to wherever "me" was pointing to, ie, the next element.  So this inserts "after" between "me" and the next linked cell.
	me.next = after
	//fmt.Printf(" after assignment: me = %+v, after = %+v\n", me, after)
}

func (me *Cell) deleteAfter() *Cell { // need to return the deleted cell.  If there is no cell after "me", panic.
	deletedCell := me.next
	if deletedCell == nil {
		panic(" no cell after me to delete.")
	}
	me.next = nil
	return deletedCell
}

func (list *linkedList) addRange(values []string) {
	// add the strings as cells to the end of the linked list, ie, append the new strings
	// First, find the end of the linked list
	var lastCell *Cell
	//fmt.Printf(" entering addRange: list = %+v\n", list)

	lastCell = list.sentinel
	//fmt.Printf(" before search: addRange.lastCell = %+v\n", lastCell)
	for lastCell.next != nil {
		lastCell = lastCell.next
	}
	//fmt.Printf(" after search: addRange.lastCell = %+v\n", lastCell)

	for _, s := range values {
		anotherCell := Cell{data: s}
		lastCell.addAfter(&anotherCell)
		lastCell = lastCell.next
	}
}

func (list *linkedList) toString(separator string) string {
	var sb strings.Builder

	sentinel := list.sentinel
	if sentinel.next == nil {
		return "" // the empty string is different from a nil string.
	}

	for cell := sentinel.next; cell != nil; cell = cell.next {
		//fmt.Printf("cell data %s, ", cell.data)
		sb.WriteString(cell.data)
		sb.WriteString(separator)
	}
	//fmt.Println()

	finalStr := sb.String()
	if finalStr == "" {
		return finalStr
	}

	if separator == "" {
		return finalStr
	}
	return finalStr[:len(finalStr)-len(separator)] // don't return the final separator
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

	var counter int // this starts as zero, so I don't intend to count the sentinel element.
	for cell := list.sentinel.next; cell != nil; cell = cell.next {
		counter++
	}

	return counter
}

func (list *linkedList) isEmpty() bool {
	return list.sentinel.next == nil
}

func (list *linkedList) push(s string) {
	lastCell := list.sentinel
	for lastCell != nil {
		lastCell = lastCell.next
	}

	anotherCell := Cell{data: s}
	lastCell.addAfter(&anotherCell)
}

func (list *linkedList) pop() string {
	if list.sentinel.next == nil {
		return ""
	}

	n := list.length()
	cell := list.sentinel.next
	for i := 0; i < n-1; i++ {
		cell = cell.next // I'm hoping this stops at the next to last element
	}
	str := cell.deleteAfter()
	return str.data
}

func main() {
	aCell := Cell{data: "Apple", next: nil}
	bCell := Cell{data: "Banana"}
	aCell.next = &bCell
	top := &aCell

	// Now to add a sentinel.  The purpose of a sentinel is to make it easy to add an item to the beginning of a linked list.  The sentinel itself never contains data,
	// just is a pointer to the next element.

	sentinel := Cell{data: "SENTINEL", next: top}
	top = &sentinel

	var counter int
	for cel := top; cel != nil; cel = cel.next {
		counter++
		fmt.Printf(" Cell.Data[%d]: %q  Next: %p\n", counter, cel.data, cel.next)
	}
}
