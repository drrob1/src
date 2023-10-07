package main

import (
	"fmt"
	"os"
	"testing"
)

var board [][]string

func TestMain(m *testing.M) { // this example is in the docs of the testing package, that I was referred to by the golang nuts google group.
	board = makeBoard(numRowsCols)
	board[3][3] = "Q"
	board[1][2] = "Q"
	dumpBoard(board)
	os.Exit(m.Run())
}

func TestSeriesIsLegal(t *testing.T) {
	fmt.Printf(" Starting TestSeriesIsLegal\n")
	good := seriesIsLegal(board, 3, 0, 0, 1)
	if good {
		fmt.Printf(" (3,0) (0,1) tested and is good.\n")
	} else {
		t.Errorf(" (3,0), (0,1) tested and should have been good, but got %t instead\n", good)
	}
	good = seriesIsLegal(board, 0, 3, 1, 0)
	if good {
		fmt.Printf(" (0,3) (1,0) tested and is good.\n")
	} else {
		t.Errorf(" (0,3) (1,0) tested and should have been good, but got %t instead\n", good)
	}

	board[3][4] = "Q"
	dumpBoard(board)

	good = seriesIsLegal(board, 3, 0, 0, 1)
	if !good {
		fmt.Printf(" (3,0) (0,1) tested and is not good as expected.\n")
	} else {
		t.Errorf(" (3,0), (0,1) tested and should have been not good, but got %t instead\n", good)
	}
	good = seriesIsLegal(board, 0, 3, 1, 0)
	if good {
		fmt.Printf(" (0,3) (1,0) tested and is good as expected.\n")
	} else {
		t.Errorf(" (0,3) (1,0) tested and should have been good, but got %t instead\n", good)
	}
	good = seriesIsLegal(board, 0, 4, 1, 0)
	if good {
		fmt.Printf(" (0,4) (1,0) tested and is good as expected.\n")
	} else {
		t.Errorf(" (0,4) (1,0) tested and should have been good, but got %t instead\n", good)
	}
}
