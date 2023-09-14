package main

import (
	"fmt"
	"math/rand"
)

// bubblesort code for Manning's live project.
// At first, I was confused that I don't have to return []int in bubbleSort.  Then I realized that slices use pointer semantics, so the result is automatically returned.

func makeRandomSlice(numItems, max int) []int {
	randomSlice := make([]int, numItems)
	if numItems < 0 {
		return nil
	}
	if max < 0 {
		return nil
	}

	for i := range randomSlice {
		randomSlice[i] = rand.Intn(max) // starting Go 1.20, don't need to init this first by calling rand.Seed()
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

func checkSorted(slice []int) {
	for i := 0; i < len(slice)-1; i++ {
		if slice[i] > slice[i+1] {
			fmt.Printf("The slice is NOT sorted.\n")
			return
		}
	}
	fmt.Printf("The slice is sorted.\n")
}

/* from wikipedia entry on buble sort.  This is pseudocode.
procedure bubbleSort(A : list of sortable items)
    n := length(A)
    repeat
        swapped := false
        for i := 1 to n-1 inclusive do
            { if this pair is out of order }
            if A[i-1] > A[i] then
                { swap them and remember something changed }
                swap(A[i-1], A[i])
                swapped := true
            end if
        end for
    until not swapped
end procedure
*/

func bubbleSort(slice []int) {
	swapped := true
	n := len(slice)

	for swapped {
		swapped = false
		for i := 1; i < n; i++ { // if this pair is out of order
			if slice[i-1] > slice[i] {
				slice[i-1], slice[i] = slice[i], slice[i-1]
				swapped = true
			}
		}
	}
}

func minInt(i, j int) int { // I don't want to clash w/ the new min functions that are built in as of Go 1.21.
	if i < j {
		return i
	}
	return j
}

func main() {

	// Get the number of items and maximum item value.
	var numItems, max int
	fmt.Printf("# Items: ")
	fmt.Scanln(&numItems)
	fmt.Printf("Max: ")
	fmt.Scanln(&max)

	// Make and display an unsorted slice.
	slice := makeRandomSlice(numItems, max)
	printSlice(slice, 40)
	fmt.Println()

	// Sort and display the result.
	bubbleSort(slice)
	printSlice(slice, 40)

	// Verify that it's sorted.
	checkSorted(slice)
}
