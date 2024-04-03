package stack

import "errors"

var ErrEmpty = errors.New("empty stack")

// Since these are generic functions, the type T needs to be specified in every function.  And then once the stack is created using a particular type, then only that
// type can be used by the other functions.  This is NOT like an empty interface, which allows several different types to be inter-mixed.
// This uses a singly linked list to be the stack.  Code provided by Miki Tebeka.

type node[T any] struct {
	value T
	next  *node[T]
}

type stackG[T any] struct {
	head *node[T]
	size int
}

func New[T any]() *stackG[T] {
	s := stackG[T]{}
	return &s
}

func (s *stackG[T]) Len() int {
	return s.size
}

func (s *stackG[T]) Push(val T) {
	s.head = &node[T]{val, s.head}
	s.size++
}

func (s *stackG[T]) Pop() (T, error) {
	if s.size == 0 {
		var zero T
		return zero, ErrEmpty
	}

	n := s.head
	s.head = n.next
	s.size--

	return n.value, nil
}

// -------------------------------------------------------------------------------------------------------------------
// From the LiveProject linked list code in Stephens_GoAlgs_LinkedLists_and_Trees_Authors_solution_1.

type Cell struct {
	data int
	next *Cell
}

type LinkedList struct {
	sentinel *Cell
}

func (me *Cell) AddAfter(after *Cell) {
	after.next = me.next
	me.next = after
}

// Delete the cell after me and return the deleted cell.

func (me *Cell) DeleteAfter() *Cell {
	// Make sure there *is* a following cell.
	if me.next == nil {
		panic("There is no cell after this one to delete")
	}

	// Return the following cell.
	after := me.next
	me.next = after.next
	return after
}

// Make a new LinkedList and initialize its sentinel.

func MakeLinkedList() LinkedList {
	list := LinkedList{}
	list.sentinel = &Cell{0, nil}
	return list
}

// Return the number of cells in the list.

func (list *LinkedList) Length() int {
	count := 0
	for cell := list.sentinel.next; cell != nil; cell = cell.next {
		count++
	}
	return count
}

// Return true if the stack is empty, false otherwise.

func (stack *LinkedList) IsEmpty() bool {
	return stack.sentinel.next == nil
}

// *** Stack functions ***

// Push an item onto the top of the list right aftr the sentinel.
func (stack *LinkedList) Push(value int) {
	cell := Cell{data: value}
	stack.sentinel.AddAfter(&cell)
}

// Pop an item off of the list (from right after the sentinel).
func (stack *LinkedList) Pop() int {
	return stack.sentinel.DeleteAfter().data
}

// -------------------------------------------------------------------------------------------------------------------
var intStack []int // for non-recursive quick sorts

type hiloIndexType struct { // for non-recursive quick sorts
	lo, hi int
}

var hiloStack []hiloIndexType // for non-recursive quick sorts

func intStackInit(n int) {
	intStack = make([]int, 0, n)
}

func intStackPush(i int) {
	intStack = append(intStack, i)
}

func intStackPop() int {
	i := intStack[len(intStack)-1]
	intStack = intStack[:len(intStack)-1]
	return i
}

func intStackLen() int {
	return len(intStack)
}

func hiloInit(n int) {
	hiloStack = make([]hiloIndexType, 0, n)
}

func hiloStackPush(i hiloIndexType) {
	hiloStack = append(hiloStack, i)
}

func hiloStackPop() hiloIndexType {
	i := hiloStack[len(hiloStack)-1]
	hiloStack = hiloStack[:len(hiloStack)-1]
	return i
}

func hiloStackLen() int {
	return len(hiloStack)
}
