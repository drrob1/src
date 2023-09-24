package main

import (
	"fmt"
	"testing"
)

func TestLength(t *testing.T) {
	queue := makeDoublyLinkedList()
	n := queue.length()
	if queue.length() != 0 {
		t.Errorf(" Length of linked list after making it should be 0, but it's %d instead.\n", n)
	}

	apple := Node{
		data:  "apple",
		left:  nil,
		right: nil,
	}
	banana := Node{
		data:  "banana",
		left:  nil,
		right: nil,
	}
	pear := Node{
		data:  "pear",
		left:  nil,
		right: nil,
	}
	peach := Node{
		data:  "peach",
		left:  nil,
		right: nil,
	}
	q := makeDoublyLinkedList()
	q.enqueue(apple)
	q.enqueue(banana)
	q.enqueue(peach)
	q.enqueue(pear)

	n = q.length()
	if n != 4 {
		t.Errorf(" Length of apple, banana, peach, pear is not 4, it is %d.\n", n)
	}
	fmt.Printf(" list is %+v\n", q.toSlice())
}

func TestToString(t *testing.T) {
	apple := Node{
		data:  "apple",
		left:  nil,
		right: nil,
	}
	banana := Node{
		data:  "banana",
		left:  nil,
		right: nil,
	}
	pear := Node{
		data:  "pear",
		left:  nil,
		right: nil,
	}
	peach := Node{
		data:  "peach",
		left:  nil,
		right: nil,
	}
	q := makeDoublyLinkedList()
	q.enqueue(apple)
	q.enqueue(banana)
	q.enqueue(pear)
	q.enqueue(peach)

	s := q.toString(", ")
	if s != "peach, pear, banana, apple" {
		leng := q.length()
		t.Errorf(" toString should have been apple, banana, pear, peach -- but it's %q instead.  Len = %d\n", s, leng)
	}

	s = q.toString(",")
	if s != "peach,pear,banana,apple" {
		leng := q.length()
		t.Errorf(" toString should have been apple,banana,pear,peach -- but it's %q instead.  Len = %d\n", s, leng)
	}

	item := q.dequeue()
	if item.data != "apple" {
		t.Errorf(" after dequeued first item is not apple.  It is %q\n", item.data)
	}
	item = q.dequeue()
	if item.data != "banana" {
		t.Errorf(" after dequeued 2nd item is not banana.  It is %q\n", item.data)
	}
	item = q.dequeue()
	if item.data != "pear" {
		t.Errorf(" after dequeued 3rd item is not pear.  It is %q\n", item.data)
	}
	item = q.dequeue()
	if item.data != "peach" {
		t.Errorf(" after dequeued last item is not peach.  It is %q\n", item.data)
	}
	if q.length() != 0 {
		t.Errorf(" Length of the queue should not be zero.  It is %d instead.\n", q.length())
	}

}

func TestToSlice(t *testing.T) {
	apple := Node{
		data:  "apple",
		left:  nil,
		right: nil,
	}
	banana := Node{
		data:  "banana",
		left:  nil,
		right: nil,
	}
	pear := Node{
		data:  "pear",
		left:  nil,
		right: nil,
	}
	peach := Node{
		data:  "peach",
		left:  nil,
		right: nil,
	}
	q := makeDoublyLinkedList()
	q.push(apple)
	q.push(banana)
	q.push(pear)
	q.push(peach)

	slice := q.toSlice()
	if slice[3] != "apple" || slice[2] != "banana" || slice[1] != "pear" || slice[0] != "peach" {
		leng := q.length()
		t.Errorf(" toSlice should have been [apple banana pear peach] but it's %+v instead.  Len = %d\n", slice, leng)
	}
}
