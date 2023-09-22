package main

import (
	"fmt"
	"testing"
)

func TestLength(t *testing.T) {
	lnkdLst := makeDoublyLinkedList()
	//fmt.Printf(" in TestLength: lnkdList = %+v\n", lnkdLst)
	//aCell := Cell{"apple", nil}
	//bCell := Cell{data: "banana"}
	//aCell.addAfter(&bCell)
	//lnkdLst.sentinel.addAfter(&aCell)
	n := lnkdLst.length()
	if lnkdLst.length() != 0 {
		t.Errorf(" Length of linked list after making it should be 0, but it's %d instead.\n", n)
	}
	lnkdLst.addRange([]string{"apple"})
	n = lnkdLst.length()
	if n != 1 {
		t.Errorf(" Length of a,b cell should be 1, but is %d instead.\n", n)
	}

	lnklst := makeDoublyLinkedList()
	lnklst.addRange([]string{"one", "two", "three"})
	n = lnklst.length()
	if n != 3 {
		t.Errorf(" Length of one, two, three is not 3, it is %d.\n", n)
	}
	fmt.Printf(" list is %+v\n", lnklst.toSlice())
}

func TestToString(t *testing.T) {
	lnkdLst := makeDoublyLinkedList()
	//lnkdLst.addRange([]string{"apple"})
	lnkdLst.addRange([]string{"apple", "banana", "pear", "peach"})
	s := lnkdLst.toString(", ")
	if s != "apple, banana, pear, peach" {
		leng := lnkdLst.length()
		t.Errorf(" toString should have been apple, banana, pear, peach -- but it's %q instead.  Len = %d\n", s, leng)
	}

	s = lnkdLst.toString(",")
	if s != "apple,banana,pear,peach" {
		leng := lnkdLst.length()
		t.Errorf(" toString should have been apple,banana,pear,peach -- but it's %q instead.  Len = %d\n", s, leng)
	}
}

func TestToSlice(t *testing.T) {
	lnkdLst := makeDoublyLinkedList()
	lnkdLst.addRange([]string{"apple", "banana", "pear", "peach"})
	slice := lnkdLst.toSlice()
	if slice[0] != "apple" || slice[1] != "banana" || slice[2] != "pear" || slice[3] != "peach" {
		leng := lnkdLst.length()
		t.Errorf(" toSlice should have been [apple banana pear peach] but it's %+v instead.  Len = %d\n", slice, leng)
	}
}
