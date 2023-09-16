package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"math/rand"
	"runtime"
	"strconv"
)

// Counting sort is an efficient algorithm for sorting an array of elements that each have a nonnegative integer key, for example,
// an array, sometimes called a list, of positive integers could have keys that are just the value of the integer as the key,
// or a list of words could have keys assigned to them by some scheme mapping the alphabet to integers (to sort in alphabetical order,
// for instance). Unlike other sorting algorithms, such as mergesort, counting sort is an integer sorting algorithm, not a comparison
// based algorithm. While any comparison based sorting algorithm requires O(n*log(n)) comparisons, counting sort has a running
// time of O(n) when the length of the input list is not much smaller than the largest key value, k, in the list.
// Counting sort can be used as a subroutine for other, more powerful, sorting algorithms such as radix sort.

type Customer struct {
	id           string
	numPurchases int
}

var verboseFlag = flag.Bool("v", false, "Verbose mode for debugging.")

const onWin = runtime.GOOS == "windows"

func countingSort(slice []Customer, max int) []Customer {
	sortedCust := make([]Customer, len(slice))
	tally := make([]int, max)

	// create the individual tallies
	for i := range slice {
		tally[slice[i].numPurchases]++
	}

	if *verboseFlag {
		fmt.Printf("After creating individual tally: %+v\n", tally)
	}

	// modify tallies so that each element includes the tallies below it
	// This step is fucked!
	for i := range slice {
		for j := 0; j < i; j++ {
			tally[i] += tally[j]
		}
	}

	if *verboseFlag {
		fmt.Printf("After creating cumulative tallies: %+v\n", tally)
	}

	// Populate output slice starting from the end
	for j := len(sortedCust) - 1; j < 0; j-- {
		tally[slice[j].numPurchases]--                      // this is a step to minimize collisions
		sortedCust[tally[slice[j].numPurchases]] = slice[j] // intentionally accessing tally-1 because this is a zero origin array.
	}

	return sortedCust
}

func main() {
	flag.Parse()

	// Get the number of items and maximum item value.
	var numItems, max int
	//fmt.Printf("# Items: ")
	//fmt.Scanln(&numItems)
	//fmt.Printf("Max: ")
	//fmt.Scanln(&max)
	numItems = 30
	max = 30

	// Make and display the unsorted slice.
	slice := makeRandomSlice(numItems, max)
	printSlice(slice, 40)
	fmt.Println()

	// Sort and display the result.
	sorted := countingSort(slice, max)
	printSlice(sorted, 40)

	// Verify that it's sorted.
	checkSorted(sorted)
}

func makeRandomSlice(numItems, max int) []Customer {
	randomSlice := make([]Customer, numItems)
	if numItems < 0 {
		return nil
	}
	if max < 0 {
		return nil
	}

	for i := range randomSlice {
		randomSlice[i].numPurchases = rand.Intn(max) // starting Go 1.20, don't need to init this first by calling rand.Seed()
		str := strconv.Itoa(i)
		randomSlice[i].id = "C" + str
	}
	return randomSlice
}

func printSlice(slice []Customer, numItems int) {
	mi := minInt(len(slice), numItems)
	pSlice := slice[:mi]
	for i, val := range pSlice {
		ctfmt.Printf(ct.Green, onWin, "%s", val.id)
		ctfmt.Printf(ct.Blue, onWin, ":")
		ctfmt.Printf(ct.Yellow, onWin, " %02d. ", val.numPurchases)
		if i%20 == 0 && i > 0 {
			ctfmt.Printf(ct.Green, onWin, " \n")
		}
	}
	fmt.Println()
}

func checkSorted(slice []Customer) {
	for i := 0; i < len(slice)-1; i++ {
		if slice[i].numPurchases > slice[i+1].numPurchases {
			fmt.Printf("The slice is NOT sorted.  That sucks!\n")
			return
		}
	}
	fmt.Printf("The slice is sorted.  Yea!\n")
}

func minInt(i, j int) int { // I don't want to clash w/ the new min functions that are built in as of Go 1.21.
	if i < j {
		return i
	}
	return j
}
