package main

import (
	"fmt"
	"os"
	"time"
)

// The Knight's Tour.  Part of the liveproject from Manning.  This is to use recursive functions.
// The open version allows the knight to start and finish on any square.
// The closed version requires the knight to finish such that it can return to its starting position in 1 more move.
// This demonstrates a class of recursive algorithms that search a solution space.  And this is a lead up to the N-queens problem.
// I just looked at the partial solution.  The author says that for an 8 x 8 board, the solution takes ~11.6 minutes if open, and > 15 min if closed.
// I need a longer timeout.  I'll try 15 min and see how it goes.
// Nope, that didn't work.  The timeout triggered before a solution did.  I'll make it 30 min and see what happens.
// I changed the initialize Offsets() to be as used by Stephens, and suddenly the 6 x 6 took ~ 1/2 sec, and the 8 x 8 took ~8 sec.  Down from > 30 min even for the 6 x 6.  Wow.
// Turns out that the order of the different moves in the Offset struct matters.  A lot!

const numRows = 8
const numCols = numRows

const requireClosedTour = false

const unvisited = -1

const timeOutMinutes = 60

type Offset struct { // legal relative moves from current position.  Positive numbers are down or right.  Must check for legal moves (not off the board or already visited).
	dr, dc int
}

var moveOffsets []Offset // all possible moves a knight can make.

var numCalls int64 // number of times the recursive function is called.

var start time.Time // I want this to be global so I can set a timeout of 10 min for now.

func initializeOffsets() { // Stephens' code.  Finds a solution in < 8 sec.
	moveOffsets = []Offset{
		Offset{-2, -1},
		Offset{-1, -2},
		Offset{+2, -1},
		Offset{+1, -2},
		Offset{-2, +1},
		Offset{-1, +2},
		Offset{+2, +1},
		Offset{+1, +2},
	}
}

func initOffsets() { // my code.  Doesn't find a solution in 1 hr.  I'm not setting the timeout for longer.
	moveOffsets = []Offset{
		Offset{dr: 2, dc: 1},
		Offset{dr: 2, dc: -1},
		Offset{dr: -2, dc: 1},
		Offset{dr: -2, dc: -1},
		Offset{dr: 1, dc: 2},
		Offset{dr: -1, dc: 2},
		Offset{dr: 1, dc: -2},
		Offset{dr: -1, dc: -2},
	}
}

func makeBoard(rows, cols int) [][]int { // I had to google this to get it right.
	aBoard := make([][]int, rows)
	for i := range aBoard {
		aBoard[i] = make([]int, cols)
	}
	for i := range aBoard[0] {
		for j := range aBoard[i] {
			aBoard[i][j] = unvisited
		}
	}
	return aBoard
}

func dumpBoard0(board [][]int) {
	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			fmt.Printf(" %02d", board[i][j])
		}
		fmt.Println()
	}
}
func dumpBoard(board [][]int) {
	for i := range board {
		for j := range board[i] {
			fmt.Printf(" %02d", board[i][j])
		}
		fmt.Println()
	}
}

func findTour(board [][]int, numRows, numCols, curRow, curCol, numVisited int) bool { // returns whether we have found a solution.
	numCalls++
	if numVisited == numRows*numCols { // every box has been visited, we're done.
		if requireClosedTour { // closed tour means that the next move of the knight is back to it's first move.  Look to see if a possible next move is the first move.
			for _, move := range moveOffsets {
				row := curRow + move.dr
				col := curCol + move.dc
				if row >= numRows || col >= numCols || row < 0 || col < 0 { // this move would be off the board.  This test prevents an out of bounds error from the next condition.
					continue
				}
				if board[row][col] == 0 { // 0 would mean the starting point
					return true
				}
			}
			return false
		} else {
			return true
		}
	}

	// Now must test if there's a legal next move
	for _, move := range moveOffsets {
		row := curRow + move.dr
		col := curCol + move.dc
		if row >= numRows || col >= numCols || row < 0 || col < 0 { // this move would be off the board.
			continue
		}
		if board[row][col] != unvisited { // skip this possible move as we've been here before.
			continue
		}
		board[row][col] = numVisited
		numVisited++
		if findTour(board, numRows, numCols, row, col, numVisited) { // this recursive call returns true when it succeeds.
			//fmt.Printf(" This move worked: row = %d, col = %d, numVisited = %d\n", row, col, numVisited)
			return true
		}
		board[row][col] = unvisited //  undo this proposed move because it didn't work.
		numVisited--                // have to undo this var because the proposed move didn't work.

		// Do we need to time out?
		elapsed := time.Since(start)
		if elapsed > timeOutMinutes*time.Minute {
			fmt.Printf(" Timeout of %d minutes triggered.  row = %d, col = %d, numVisited = %d.  NumCalls = %d  Elapsed = %s. Exiting.\n",
				timeOutMinutes, row, col, numVisited, numCalls, elapsed.String())
			os.Exit(1)
		}
	}

	return false
}

func oldmain() {
	//initializeOffsets()
	board := makeBoard(numRows, numCols)
	dumpBoard0(board)
	board[4][4] = 1
	board[4][5] = 2
	board[3][2] = 3
	board[1][1]++
	//fmt.Printf("%+v\n", board)
	fmt.Println()
	dumpBoard(board)

}

func init() {
	//initOffsets()
	initializeOffsets()
}

func main() {
	oldmain()
	fmt.Printf("\n\n\n")

	numCalls = 0

	// Initialize the move offsets.
	//initializeOffsets()

	// Create the blank board.
	board := makeBoard(numRows, numCols)

	// Try to find a tour.
	start = time.Now()
	board[0][0] = 0 // this is the starting point.
	if findTour(board, numRows, numCols, 0, 0, 1) {
		fmt.Println("Success!")
	} else {
		fmt.Println("Could not find a tour.")
	}
	elapsed := time.Since(start)
	dumpBoard(board)
	fmt.Printf("%f seconds, and %s\n", elapsed.Seconds(), elapsed.String())
	fmt.Printf("%d calls\n", numCalls)
}
