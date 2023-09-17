package main

import (
	"fmt"
	"math/rand"
	"strconv"
)

const maxPrinted = 40

func binarySearch(slice []int, target int) (int, int) {
	var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		numTries++
		if slice[current] < target {
			left = current + 1
		} else if slice[current] > target {
			right = current - 1
		} else {
			return current, numTries
		}
	}
	return -1, numTries
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

	// quicksort and display the slice.
	quicksort(slice)
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
		index, numTests := binarySearch(slice, target)
		if index >= 0 {
			fmt.Printf(" Found target of %d at position %d, using %d tests.\n", target, index, numTests)
		} else {
			fmt.Printf(" Target of %d not found after %d tries.\n", target, numTests)
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
func quicksort(slice []int) {
	lo := 0
	hi := len(slice) - 1

	partition(slice, lo, hi) // get pivot index and partition the slice
}

func partition(slice []int, lo, hi int) {
	i := lo
	j := hi
	pivot := slice[(lo+hi)/2]
	for i <= j {
		for slice[i] < pivot {
			i++
		}
		for pivot < slice[j] {
			j--
		}
		if i <= j {
			slice[i], slice[j] = slice[j], slice[i]
			i++
			j--
		}
	}
	if lo < j {
		partition(slice, lo, j)
	}
	if i < hi {
		partition(slice, i, hi)
	}
	return
}
