/*
  17 Sep 23 -- Singly linked list for the live project.
  19 Sep 23 -- Added hasLoop and toStringMax for the next part of this project.
  19 Sep 23 -- Now called doublylinkedlist.
                 It was hard for me to write addAfter.  I decided to write what I have and then what I need, and then it was easy for me to figure out how to get from input to output.
                 Input: me.next points to right.  Right.prev points to me.  I renamed me to left.  I get the right cell from initial left.next.  I don't change left.prev or right.next.  I only have to change left.next and right.prev.
                 Output: btwn.next becomes initial left.next, btwn.prev = left, left.next = btwn, right.prev = btwn.
               A double linked list is a linear thing.  A tree is non-linear.
*/

package main

import (
	"fmt"
	"strings"
)

type Cell struct {
	data       string
	prev, next *Cell
}

type doublyLinkedList struct {
	topSentinel, bottomSentinel *Cell
}

func makeDoublyLinkedList() doublyLinkedList {
	// This is a factory function to create a new doubly linked list, init the sentinel pointers to a new cell, and return the linked list.
	list := doublyLinkedList{}
	fcell := Cell{data: "FrontSentinal"}
	bcell := Cell{data: "BackSentinal"}
	fcell.next = &bcell
	bcell.prev = &fcell
	list.topSentinel = &fcell
	list.bottomSentinel = &bcell
	// fmt.Printf(" in makeLinkedList.  List: %+v\n", list)
	return list
}

func (left *Cell) addAfter(btwn *Cell) {
	right := left.next // right now is a pointer to the initial right hand cell, so btwn can be inserted in btwn left and right cells.
	btwn.next = right  // so the next field of "btwn" now points to wherever "left" was pointing to, ie, the element after btwn that I'm calling "right".  So this inserts "btwn" between "left" and "right".
	btwn.prev = left
	left.next = btwn
	right.prev = btwn
	// no change to me.prev and right.next
	// fmt.Printf(" after assignment: me = %+v, btwnCell = %+v, right = %+v\n", btwn, btwn, right)
}

func (right *Cell) addBefore(btwn *Cell) {
	left := right.prev
	left.addAfter(btwn)
}

func (me *Cell) delete() { // need to return the deleted cell.  If there is no cell after "me", panic.
	//fmt.Printf(" in delete.  me.data = %q\n", me.data)
	left := me.prev
	right := me.next
	left.next = right
	right.prev = left
}

func (list *doublyLinkedList) addRange(values []string) {
	// add the strings as cells to the end of the linked list, ie, append the new strings

	for _, s := range values {
		addMe := Cell{data: s}
		list.bottomSentinel.addBefore(&addMe)
	}
}

func (list *doublyLinkedList) toString(separator string) string {
	var sb strings.Builder

	if list.isEmpty() {
		return ""
	}

	for cell := list.topSentinel.next; cell != list.bottomSentinel; cell = cell.next {
		//fmt.Printf("cell data %s, ", cell.data)
		sb.WriteString(cell.data)
		if cell.next != list.bottomSentinel { // I got this solution from the course hint.
			sb.WriteString(separator)
		}
	}
	//fmt.Println()
	return sb.String()
}

func (list *doublyLinkedList) toSlice() []string {
	var stringSlice []string

	if list.isEmpty() {
		return []string{}
	}

	for cell := list.topSentinel.next; cell != list.bottomSentinel; cell = cell.next {
		stringSlice = append(stringSlice, cell.data)
	}

	return stringSlice
}

func (list *doublyLinkedList) length() int {
	if list.topSentinel.next == list.bottomSentinel {
		return 0
	}

	var counter int // this starts as zero, so I don't intend to count the sentinel elements.
	for cell := list.topSentinel.next; cell != list.bottomSentinel; cell = cell.next {
		counter++
	}

	return counter
}

func (list *doublyLinkedList) isEmpty() bool {
	return list.topSentinel.next == list.bottomSentinel
}

func (list *doublyLinkedList) push(s string) { // I'm going to change this so push and pop occur at the top of the list
	anotherCell := Cell{data: s}

	list.topSentinel.addAfter(&anotherCell)
}

func (list *doublyLinkedList) pop() string { // I'm going to change this so push and pop occur at the top of the list
	cell := list.topSentinel.next
	//fmt.Printf(" in pop: Cell = %q\n", cell.data)
	cell.delete()
	return cell.data
}

func (list *doublyLinkedList) hasLoop() bool {
	// This uses a fast and slow pointer to Cell.  Fast moves two elements for each one that slow moves.  The loop terminates if a nil is found, or fast = slow.
	// Must move fast first, check for match w/ slow, and then move slow.
	var fast, slow *Cell

	fast = list.topSentinel
	slow = fast

	for {
		fast = fast.next
		if fast.next == list.bottomSentinel {
			return false
		}
		if fast.data == slow.data && fast.next == slow.next {
			return true
		}
		fast = fast.next
		if fast.next == list.bottomSentinel {
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

func (list *doublyLinkedList) toStringMax(sep string, maxx int) string { // check maxx number of cells.  This is make sure it will stop if there's a loop.
	var sb strings.Builder
	var counter int

	if list.topSentinel.next == list.bottomSentinel {
		return ""
	}

	for cell := list.topSentinel.next; cell != list.bottomSentinel; cell = cell.next {
		//fmt.Printf("cell data %s, ", cell.data)
		sb.WriteString(cell.data)
		if cell.next != list.bottomSentinel { // I got this solution from the course hint.
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

func (queue *doublyLinkedList) enqueue(value string) {
	queue.push(value)
}

func (queue *doublyLinkedList) dequeue() string {
	cell := queue.bottomSentinel.prev
	//fmt.Printf(" in pop: Cell = %q\n", cell.data)
	cell.delete()
	return cell.data
}

func (deque *doublyLinkedList) pushBottom(value string) {
	anotherCell := Cell{data: value}

	deque.bottomSentinel.addBefore(&anotherCell)
}

func (deque *doublyLinkedList) pushTop(value string) {
	deque.push(value)
}

func (deque *doublyLinkedList) popTop() string {
	return deque.pop()
}

func (deque *doublyLinkedList) popBottom() string {
	return deque.dequeue()
}

func oldmain() {
	// smallListTest()

	// Make a list from an array of values.
	greekLetters := []string{
		"α", "β", "γ", "δ", "ε",
	}
	list1 := makeDoublyLinkedList()
	list1.addRange(greekLetters)
	fmt.Println(list1.toString(" "))
	fmt.Println()

	// Demonstrate a stack.
	fmt.Printf(" Stack operations\n")
	stack := makeDoublyLinkedList()
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
	list := makeDoublyLinkedList()
	list.addRange(values)

	fmt.Println(list.toString(" "))
	if list.hasLoop() {
		fmt.Println("Has loop")
	} else {
		fmt.Println("No loop")
	}
	fmt.Println()

	// Make cell 5 point to cell 2.
	list.topSentinel.next.next.next.next.next.next = list.topSentinel.next.next

	fmt.Println(list.toStringMax(" ", 10))
	if list.hasLoop() {
		fmt.Println("Has loop")
	} else {
		fmt.Println("No loop")
	}
	fmt.Println()

	// Make cell 4 point to cell 2.
	list.topSentinel.next.next.next.next.next = list.topSentinel.next.next

	fmt.Println(list.toStringMax(" ", 10))
	if list.hasLoop() {
		fmt.Println("Has loop")
	} else {
		fmt.Println("No loop")
	}

	fmt.Printf("\n\n\n")
}

func main() {
	oldmain()

	// Test queue functions.
	fmt.Printf("*** Queue Functions ***\n")
	queue := makeDoublyLinkedList()
	queue.enqueue("Agate")
	queue.enqueue("Beryl")
	fmt.Printf("%s ", queue.dequeue())
	queue.enqueue("Citrine")
	fmt.Printf("%s ", queue.dequeue())
	fmt.Printf("%s ", queue.dequeue())
	queue.enqueue("Diamond")
	queue.enqueue("Emerald")
	for !queue.isEmpty() {
		fmt.Printf("%s ", queue.dequeue())
	}
	fmt.Printf("\n\n")

	// Test deque functions. Names starting
	// with F have a fast pass.
	fmt.Printf("*** Deque Functions ***\n")
	deque := makeDoublyLinkedList()
	deque.pushTop("Ann")
	deque.pushTop("Ben")
	fmt.Printf("%s ", deque.popBottom())
	deque.pushBottom("F-Cat")
	fmt.Printf("%s ", deque.popBottom())
	fmt.Printf("%s ", deque.popBottom())
	deque.pushBottom("F-Dan")
	deque.pushTop("Eva")
	for !deque.isEmpty() {
		fmt.Printf("%s ", deque.popBottom())
	}
	fmt.Printf("\n")
}
