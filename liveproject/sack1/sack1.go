package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
   The first part of the knapsack problem for the Manning live project by Rod Stephens.  It uses an exhaustive search approach.

*/

const numItems = 20 // A reasonable value for exhaustive search.

const minValue = 1
const maxValue = 10
const minWeight = 4
const maxWeight = 10

var allowedWeight int

type Item struct {
	value, weight int
	isSelected    bool
}

// Make some random items.
func makeItems(numItems, minValue, maxValue, minWeight, maxWeight int) []Item {
	// Initialize a pseudorandom number generator.
	//random := rand.New(rand.NewSource(time.Now().UnixNano())) // Initialize with a changing seed  Not needed as of Go 1.20
	random := rand.New(rand.NewSource(1337)) // Initialize with a fixed seed  Not needed as of Go 1.20

	items := make([]Item, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = Item{
			random.Intn(maxValue-minValue+1) + minValue,
			random.Intn(maxWeight-minWeight+1) + minWeight,
			false}
	}
	return items
}

// Return a copy of the items slice.
func copyItems(items []Item) []Item {
	newItems := make([]Item, len(items))
	copy(newItems, items)
	return newItems
}

// Return the total value of the items.
// If addAll is false, only add up the selected items.
func sumValues(items []Item, addAll bool) int {
	total := 0
	for i := 0; i < len(items); i++ {
		if addAll || items[i].isSelected {
			total += items[i].value
		}
	}
	return total
}

// Return the total weight of the items.
// If addAll is false, only add up the selected items.
func sumWeights(items []Item, addAll bool) int {
	total := 0
	for i := 0; i < len(items); i++ {
		if addAll || items[i].isSelected {
			total += items[i].weight
		}
	}
	return total
}

// Return the value of this solution.
// If the solution is too heavy, return -1 so we prefer an empty solution.
func solutionValue(items []Item, allowedWeight int) int {
	// If the solution's total weight > allowedWeight,
	// return -1 so we won't use this solution.
	if sumWeights(items, false) > allowedWeight {
		return -1
	}

	// Return the sum of the selected values.
	return sumValues(items, false)
}

// Print the selected items.
func printSelected(items []Item) {
	numPrinted := 0
	for i, item := range items {
		if item.isSelected {
			fmt.Printf("%d(%d, %d) ", i, item.value, item.weight)
		}
		numPrinted += 1
		if numPrinted > 100 {
			fmt.Println("...")
			return
		}
	}
	fmt.Println()
}

// Run the algorithm. Display the elapsed time and solution.
func runAlgorithm(alg func([]Item, int) ([]Item, int, int), items []Item, allowedWeight int) {
	// Copy the items so the run isn't influenced by a previous run.
	testItems := copyItems(items)

	start := time.Now()

	// Run the algorithm.
	solution, totalValue, functionCalls := alg(testItems, allowedWeight)

	elapsed := time.Since(start)

	fmt.Printf("Elapsed: %f sec, %s\n", elapsed.Seconds(), elapsed.String())
	printSelected(solution)
	fmt.Printf("Value: %d, Weight: %d, Calls: %d\n",
		totalValue, sumWeights(solution, false), functionCalls)
	fmt.Println()
}

// Recursively assign values in or out of the solution.
// Return the best assignment, value of that assignment,
// and the number of function calls we made.
func exhaustiveSearch(items []Item, allowedWeight int) ([]Item, int, int) {
	return doExhaustiveSearch(items, allowedWeight, 0)
}

func doExhaustiveSearch(items []Item, allowedWeight, nextIndex int) ([]Item, int, int) {
	currentValue := solutionValue(items, allowedWeight)
	if nextIndex >= len(items) || currentValue < 0 {
		copyOfItems := copyItems(items)
		return copyOfItems, currentValue, 1
	}

	// have not reached the bottom of the solution space tree and have not exceeded the constraint.
	items[nextIndex].isSelected = true
	resultSolnIfYes, totalValueIfYes, numberOfFunctionCallsIfYes := doExhaustiveSearch(items, allowedWeight, nextIndex+1)
	//fmt.Printf(" doing exhaustive search: totalValueIfYes = %d, allowedWeight=%d\n", totalValueIfYes, allowedWeight)
	items[nextIndex].isSelected = false
	resultSolnIfNo, totalValueIfNo, numberOfFunctionCallsIfNo := doExhaustiveSearch(items, allowedWeight, nextIndex+1)
	//fmt.Printf(" doing exhaustive search: totalValueIfNo = %d, allowedWeight=%d\n", totalValueIfNo, allowedWeight)
	if totalValueIfYes >= totalValueIfNo && totalValueIfYes > 0 {
		return resultSolnIfYes, totalValueIfYes, numberOfFunctionCallsIfYes + numberOfFunctionCallsIfNo + 1
	}
	return resultSolnIfNo, totalValueIfNo, numberOfFunctionCallsIfNo + numberOfFunctionCallsIfYes + 1
}

func main() {
	//items := makeTestItems()
	items := makeItems(numItems, minValue, maxValue, minWeight, maxWeight)
	allowedWeight = sumWeights(items, true) / 2

	// Display basic parameters.
	fmt.Println("*** Parameters ***")
	fmt.Printf("# items: %d\n", numItems)
	fmt.Printf("Total value: %d\n", sumValues(items, true))
	fmt.Printf("Total weight: %d\n", sumWeights(items, true))
	fmt.Printf("Allowed weight: %d\n", allowedWeight)
	fmt.Println()

	// Exhaustive search
	if numItems > 65 { // Only run exhaustive search if numItems <= 23.
		fmt.Println("Too many items for exhaustive search")
	} else {
		fmt.Println("*** Exhaustive Search ***")
		runAlgorithm(exhaustiveSearch, items, allowedWeight)
	}
}
