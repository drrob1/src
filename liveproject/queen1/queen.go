package main

import (
	"fmt"
	"strings"
	"time"
)

// The 8 Queens Puzzle from the Go project I'm doing now.  The goal is placing the 8 queens on a std 8 x 8 chessboard so that no queens threaten one another.
// For this size board, there are 12 distinct solutions.  The total of 92 solutions can be reduced by eliminating the duplicates that are derived from the fundamental solution by
// a rotation or reflection operation.
// Here we'll assume a square board.

const numRowsCols = 5 // to start he wants 5.

func makeBoard(rows int) [][]string { // I had to google this to get it right.
	aBoard := make([][]string, rows)
	for i := range aBoard {
		aBoard[i] = make([]string, rows)
	}

	for i := range aBoard[0] {
		for j := range aBoard[i] {
			aBoard[i][j] = "."
		}
	}
	return aBoard
}

func dumpBoard(board [][]string) {
	fmt.Println()
	for i := range board {
		for j := range board[i] {
			fmt.Printf("%s ", board[i][j])
		}
		fmt.Println()
	}
	fmt.Println()
}

func seriesIsLegal(board [][]string, r0, c0, dr, dc int) bool {
	numRows := len(board)
	var numQueens int

	r := r0
	c := c0

	for {
		if r >= numRows || c >= numRows {
			break
		}
		if strings.ToUpper(board[r][c]) == "Q" {
			numQueens++
		}
		r += dr
		c += dc
	}
	if numQueens < 2 {
		return true
	}
	return false
}

func boardIsLegal(board [][]string) bool {
	numRows := len(board)

	// check rows
	for r := range board {
		good := seriesIsLegal(board, r, 0, 1, 0)
		if !good {
			return false
		}
	}

	// check cols
	for c := range board {
		good := seriesIsLegal(board, 0, c, 0, 1)
		if !good {
			return false
		}
	}

	// check diagonals.  I have to check 4 diagonal loops
	for r := range board {
		good := seriesIsLegal(board, r, 0, 1, 1) // go down
		if !good {
			return false
		}
	}
	for c := range board {
		good := seriesIsLegal(board, 0, c, 1, 1) // go across
		if !good {
			return false
		}
	}
	for r := numRows - 1; r >= 0; r-- {
		good := seriesIsLegal(board, r, 0, 1, -1) // go down starting from the right
		if !good {
			return false
		}
	}
	for c := numRows - 1; c >= 0; c-- {
		good := seriesIsLegal(board, 0, c, 1, -1) // go across starting from the right
		if !good {
			return false
		}
	}

	return true
}

func main() {
	// const numRows = 5
	board := makeBoard(numRowsCols)
	board[3][3] = "Q"
	board[1][2] = "Q"
	dumpBoard(board)

	start := time.Now()
	//success := placeQueens1(board, numRowsCols, 0, 0)
	//success := placeQueens2(board, numRowsCols, 0, 0, 0)
	//success := placeQueens3(board, numRowsCols, 0, 0, 0)

	elapsed := time.Since(start)
	//if success {
	//	fmt.Println("Success!")
	//	dumpBoard(board)
	//} else {
	//	fmt.Println("No solution")
	//}
	fmt.Printf("Elapsed: %f seconds\n", elapsed.Seconds())
}
