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

func TestBoardIsLegal(t *testing.T) {
	board[3][4] = "."

	dumpBoard(board)
	good := boardIsLegal(board)
	if good {
		fmt.Printf(" Board is legal, which is expected.\n")
	} else {
		t.Errorf(" Expected board is legal, but got %t instead.\n", good)
	}

	board[2][1] = "Q"
	dumpBoard(board)
	good = boardIsLegal(board)
	if !good {
		fmt.Printf(" Board is not legal, which is expected.\n")
	} else {
		t.Errorf(" Expected board to not be legal, but it's %t instead.\n", good)
	}

	board[2][1] = "."
	board[2][4] = "q"
	dumpBoard(board)
	good = boardIsLegal(board)
	if !good {
		fmt.Printf(" Board is not legal, which is expected.\n")
	} else {
		t.Errorf(" Expected board to not be legal, but it's %t instead.\n", good)
	}

	board[2][4] = "."
	board[4][4] = "Q"
	dumpBoard(board)
	good = boardIsLegal(board)
	if !good {
		fmt.Printf(" Board is not legal, which is expected.\n")
	} else {
		t.Errorf(" Expected board to not be legal, but it's %t instead.\n", good)
	}

	board[4][4] = "."
	board[0][1] = "Q"
	dumpBoard(board)
	good = boardIsLegal(board)
	if !good {
		fmt.Printf(" Board is not legal, which is expected.\n")
	} else {
		t.Errorf(" Expected board to not be legal, but it's %t instead.\n", good)
	}

	board = makeBoard(numRowsCols)
	board[2][0] = "Q"
	board[3][1] = "q"
	dumpBoard(board)
	good = boardIsLegal(board)
	if !good {
		fmt.Printf(" Board is not legal, which is expected.\n")
	} else {
		t.Errorf(" Expected board to not be legal, but it's %t instead.\n", good)
	}

	board = makeBoard(8)
	board[5][5] = "Q"
	dumpBoard(board)
	good = boardIsLegal(board)
	if good {
		fmt.Printf(" Big board is legal, which is expected.\n")
	} else {
		t.Errorf(" Expected big board to be legal, but it's %t instead.\n", good)
	}
}
