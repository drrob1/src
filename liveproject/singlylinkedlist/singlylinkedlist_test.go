package main

import (
	"testing"
)

func TestLength(t *testing.T) {
	lnkdLst := makeLinkedList()
	//fmt.Printf(" in TestLength: lnkdList = %+v\n", lnkdLst)
	//aCell := Cell{"apple", nil}
	//bCell := Cell{data: "banana"}
	//aCell.addAfter(&bCell)
	//lnkdLst.sentinel.addAfter(&aCell)
	lnkdLst.addRange([]string{"apple"})
	n := lnkdLst.length()
	if n != 1 {
		t.Errorf(" Length of a,b cell should be 2, but is %d instead.\n", n)
	}

	lnklst := makeLinkedList()
	lnklst.addRange([]string{"one", "two", "three"})
	n = lnklst.length()
	if n != 3 {
		t.Errorf(" Length of one, two, three is not 4, it is %d.\n", n)
	}
}
