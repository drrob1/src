/*
  17 Sep 23 -- Singly linked list for the live project.
  19 Sep 23 -- Added hasLoop and toStringMax for the next part of this project.
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
	//fmt.Printf(" in deleteAfter.  me.data = %q\n", me.data)
	deletedCell := me.next
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
		if cell.next != nil { // I got this solution from the course hint.
			sb.WriteString(separator)
		}
	}
	//fmt.Println()
	return sb.String()

	//finalStr := sb.String()
	//if finalStr == "" {
	//	return finalStr
	//}
	//
	//if separator == "" {
	//	return finalStr
	//}
	//return finalStr[:len(finalStr)-len(separator)] // don't return the final separator
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
	for lastCell.next != nil {
		lastCell = lastCell.next
	}

	anotherCell := Cell{data: s}
	lastCell.addAfter(&anotherCell)
}

func (list *linkedList) pop() string {
	cell := list.sentinel
	//fmt.Printf(" in pop: Cell = %q\n", cell.data)
	n := list.length()
	if n > 0 {
		for i := 0; i < n-1; i++ {
			cell = cell.next
		}
	}
	str := cell.deleteAfter()
	return str.data
}

//func main() {
//	aCell := Cell{data: "Apple", next: nil}
//	bCell := Cell{data: "Banana"}
//	aCell.next = &bCell
//	top := &aCell
//
//	// Now to add a sentinel.  The purpose of a sentinel is to make it easy to add an item to the beginning of a linked list.  The sentinel itself never contains data,
//	// just is a pointer to the next element.
//
//	sentinel := Cell{data: "SENTINEL", next: top}
//	top = &sentinel
//
//	var counter int
//	for cel := top; cel != nil; cel = cel.next {
//		counter++
//		fmt.Printf(" Cell.Data[%d]: %q  Next: %p\n", counter, cel.data, cel.next)
//	}
//}

func (list *linkedList) hasLoop() bool {
	// This uses a fast and slow pointer to Cell.  Fast moves two elements for each one that slow moves.  The loop terminates if a nil is found, or fast = slow.
	// Must move fast first, check for match w/ slow, and then move slow.
	var fast, slow *Cell

	fast = list.sentinel
	slow = list.sentinel

	for {
		fast = fast.next
		if fast.next == nil {
			return false
		}
		if fast.data == slow.data && fast.next == slow.next {
			return true
		}
		fast = fast.next
		if fast.next == nil {
			return false
		}
		if fast.data == slow.data && fast.next == slow.next {
			return true
		}
		slow = slow.next
		if fast.data == slow.data && fast.next == slow.next {
			return true
		}
	}
}

func (list *linkedList) toStringMax(sep string, maxx int) string { // check maxx number of cells.  This is make sure it will stop if there's a loop.
	var sb strings.Builder
	var counter int

	sentinel := list.sentinel
	if sentinel.next == nil {
		return "" // the empty string is different from a nil string.
	}

	for cell := sentinel.next; cell != nil; cell = cell.next {
		//fmt.Printf("cell data %s, ", cell.data)
		sb.WriteString(cell.data)
		if cell.next != nil { // I got this solution from the course hint.
			sb.WriteString(sep)
		}
		counter++
		if counter >= maxx {
			break
		}
	}
	//fmt.Println()
	return sb.String()
}

func main() {
	// smallListTest()

	// Make a list from an array of values.
	greekLetters := []string{
		"α", "β", "γ", "δ", "ε",
	}
	list1 := makeLinkedList()
	list1.addRange(greekLetters)
	fmt.Println(list1.toString(" "))
	fmt.Println()

	// Demonstrate a stack.
	fmt.Printf(" Stack operations\n")
	stack := makeLinkedList()
	stack.push("Apple")

	stack.push("Banana")
	stack.push("Coconut")
	stack.push("Date")
	for !stack.isEmpty() {
		fmt.Printf("Popped: %-7s   Remaining %d: %s\n",
			stack.pop(),
			stack.length(),
			stack.toString(" "))
	}

	fmt.Printf("\n Now to test the looping stuff.\n\n")

	// Make a list from an array of values.
	values := []string{
		"0", "1", "2", "3", "4", "5",
	}
	list := makeLinkedList()
	list.addRange(values)

	fmt.Println(list.toString(" "))
	if list.hasLoop() {
		fmt.Println("Has loop")
	} else {
		fmt.Println("No loop")
	}
	fmt.Println()

	// Make cell 5 point to cell 2.
	list.sentinel.next.next.next.next.next.next = list.sentinel.next.next

	fmt.Println(list.toStringMax(" ", 10))
	if list.hasLoop() {
		fmt.Println("Has loop")
	} else {
		fmt.Println("No loop")
	}
	fmt.Println()

	// Make cell 4 point to cell 2.
	list.sentinel.next.next.next.next.next = list.sentinel.next.next

	fmt.Println(list.toStringMax(" ", 10))
	if list.hasLoop() {
		fmt.Println("Has loop")
	} else {
		fmt.Println("No loop")
	}
}
