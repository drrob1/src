package main

import (
	"fmt"
	"math/rand"
	"strconv"
)

const maxPrinted = 40

func linearSearch(slice []int, target int) (int, int) {
	for i, v := range slice {
		if target == v {
			return i, i + 1
		}
	}
	return -1, len(slice)
}

func main() {
	// Get the number of items and maximum item value.
	var numItems, maxx int
	fmt.Printf("# Items: ")
	fmt.Scanln(&numItems)
	fmt.Printf("Max: ")
	fmt.Scanln(&maxx)

	// Make and display the unsorted slice.
	slice := makeRandomSlice(numItems, maxx)
	printSlice(slice, maxPrinted)
	fmt.Println()

	for {
		var ans string
		fmt.Printf(" Enter a target value: ")
		fmt.Scanln(&ans)
		if ans == "" {
			break
		}
		target, err := strconv.Atoi(ans)
		if err != nil {
			fmt.Printf(" ERROR: %s.  Exiting.\n", err)
			break
		}
		index, numTests := linearSearch(slice, target)
		if index >= 0 {
			fmt.Printf(" Found target of %d at position %d, using %d tests.\n", target, index, numTests)
		} else {
			fmt.Printf(" Target of %d not found.\n", target)
		}
		fmt.Println()
	}
}

func makeRandomSlice(numItems, max int) []int {
	randomSlice := make([]int, numItems)
	if numItems < 0 {
		return nil
	}
	if max < 0 {
		return nil
	}

	for i := range randomSlice {
		randomSlice[i] = rand.Intn(max) // starting Go 1.20, rand.Seed() is deprecated.
	}
	return randomSlice
}

func printSlice(slice []int, numItems int) {
	mi := minInt(len(slice), numItems)
	pSlice := slice[:mi]
	for _, val := range pSlice {
		fmt.Printf("%d ", val)
	}
	fmt.Println()
}

func minInt(i, j int) int { // I don't want to clash w/ the new min functions that are built in as of Go 1.21.
	if i < j {
		return i
	}
	return j
}
