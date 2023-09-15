package main

import (
	"fmt"
	"math/rand"
)

// quicksort code for Manning's liveproject

// Many implementations select the first or last element as the initial pivot.  In the partitioning step,  all elements below the pivot go to the front of the slice,
// and all elements above the pivot go to the end of the slice.  Doing this quickly without using a lot of extra memory is the tricky part of the algorithm.

const maxPrinted = 70

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

	// Sort and display the result.
	quicksort(slice)
	printSlice(slice, maxPrinted)

	// Verify that it's sorted.
	checkSorted(slice)

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

func minInt(i, j int) int { // I don't want to clash w/ the new min functions that are built in as of Go 1.21.
	if i < j {
		return i
	}
	return j
}

/*

Lomuto partition scheme

This scheme is attributed to Nico Lomuto and popularized by Bentley in his book Programming Pearls[12] and Cormen et al. in their book Introduction to Algorithms.[13] In most formulations this scheme chooses as the pivot the last element in the array. The algorithm maintains index i as it scans the array using another index j such that the elements at lo through i-1 (inclusive) are less than the pivot, and the elements at i through j (inclusive) are equal to or greater than the pivot. As this scheme is more compact and easy to understand, it is frequently used in introductory material, although it is less efficient than Hoare's original scheme e.g., when all elements are equal.[14] The complexity of Quicksort with this scheme degrades to O(n2) when the array is already in order, due to the partition being the worst possible one.[10] There have been various variants proposed to boost performance including various ways to select the pivot, deal with equal elements, use other sorting algorithms such as insertion sort for small arrays, and so on. In pseudocode, a quicksort that sorts elements at lo through hi (inclusive) of an array A can be expressed as:[13]

// Sorts a (portion of an) array, divides it into partitions, then sorts those
algorithm quicksort(A, lo, hi) is
  // Ensure indices are in correct order
  if lo >= hi || lo < 0 then
    return

  // Partition array and get the pivot index
  p := partition(A, lo, hi)

  // Sort the two partitions
  quicksort(A, lo, p - 1) // Left side of pivot
  quicksort(A, p + 1, hi) // Right side of pivot

// Divides array into two partitions
algorithm partition(A, lo, hi) is
  pivot := A[hi] // Choose the last element as the pivot

  // Temporary pivot index
  i := lo - 1

  for j := lo to hi - 1 do
    // If the current element is less than or equal to the pivot
    if A[j] <= pivot then
      // Move the temporary pivot index forward
      i := i + 1
      // Swap the current element with the element at the temporary pivot index
      swap A[i] with A[j]

  // Move the pivot element to the correct pivot position (between the smaller and larger elements)
  i := i + 1
  swap A[i] with A[hi]
  return i // the pivot index

I couldn't get this to work.
*/
