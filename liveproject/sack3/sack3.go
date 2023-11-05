package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

/*
   The first part of the knapsack problem for the Manning live project by Rod Stephens.  It uses an exhaustive search approach.

   The 2nd part of the knapsack problem.  This version will used branch and bound to reduce the solution space.  The proposed solution branch of the tree is checked against
   estimated bounds, and is discarded if it doesn't produce better results than the best one found so far.
   Here, we're going to use bounds on weight and value.
   Outline
   bestValue = value of the best solution found so far
   currentValue = hmm, current solution's total value
   currentWeight = well, guess
   remainingValue = remaining total value available, initially the sum of all item values.
   If there's a complete assignment, return it
   If not, check the value bound, currentValue + remainingValue > bestValue.  If we can't improve on bestValue, return nil
   If not bailed, check the weight bound.  This is more complex, so the code will show it.

   The 3rd part of the knapsack problem.  This will use his refinement to the branch and bound reduction of the solution space.
   It involves one item blocking another item.

*/

const numItems = 20 // A reasonable value for exhaustive search.

const minValue = 1
const maxValue = 10
const minWeight = 4
const maxWeight = 10

var allowedWeight int

type Item struct {
	id, blockedBy int
	blockList     []int
	value, weight int
	isSelected    bool
}

// Make some random items.
func makeItems(numItems, minValue, maxValue, minWeight, maxWeight int) []Item {
	// Initialize a pseudorandom number generator.
	//random := rand.New(rand.NewSource(time.Now().UnixNano())) // Initialize with a changing seed  Not needed as of Go 1.20
	//random := rand.New(rand.NewSource(1337)) // Initialize with a fixed seed  Not needed as of Go 1.20

	items := make([]Item, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = Item{
			id:         i,
			blockedBy:  -1,
			blockList:  nil,
			value:      rand.Intn(maxValue-minValue+1) + minValue,
			weight:     rand.Intn(maxWeight-minWeight+1) + minWeight,
			isSelected: false}
	}
	return items
}

func makeBlockLists(items []Item) {
	for i := range items {
		items[i].blockList = []int{}
		for j := range items {
			if i != j && items[i].value >= items[j].value && items[i].weight <= items[j].weight {
				items[i].blockList = append(items[i].blockList, j)
			}
		}
	}
}

func blockItems(source Item, items []Item) {
	for i := range source.blockList {
		if items[i].blockedBy == -1 {
			items[i].blockedBy = source.id
		}
	}
}

func unBlockItems(source Item, items []Item) {
	for i := range source.blockList {
		if items[i].blockedBy == source.id {
			items[i].blockedBy = -1
		}
	}
}

func RodsTechnique(items []Item, allowedWeight int) ([]Item, int, int) {
	bestValue := 0
	currentValue := 0
	currentWeight := 0
	remainingValue := sumValues(items, true)
	makeBlockLists(items)
	return doRodsTechnique(items, allowedWeight, 0, bestValue, currentValue, currentWeight, remainingValue)
}

func RodsTechniqueSorted(items []Item, allowedWeight int) ([]Item, int, int) {
	makeBlockLists(items)
	lessFcn := func(i, j int) bool {
		return len(items[i].blockList) > len(items[j].blockList)
	}
	sort.Slice(items, lessFcn)
	// reset the IDs of each item
	for i := range items {
		items[i].id = i
	}
	// rebuild the lists w/ the new indices
	makeBlockLists(items)

	bestValue := 0
	currentValue := 0
	currentWeight := 0
	remainingValue := sumValues(items, true)
	return doRodsTechnique(items, allowedWeight, 0, bestValue, currentValue, currentWeight, remainingValue)
}

func doRodsTechnique(items []Item, allowedWeight, nextIndex, bestValue, currentValue, currentWeight, remainingValue int) ([]Item, int, int) {
	var test1Solution, test2Solution []Item
	var test1Value, test1Calls, test2Value, test2Calls int

	if nextIndex >= len(items) {
		copyOfItems := copyItems(items)
		return copyOfItems, currentValue, 1
	}

	// we do not have a full assignment.  Can we improve on this solution so it's worth continuing
	if currentValue+remainingValue <= bestValue {
		// No, we can't improve on the best solution found so far
		return nil, currentValue, 1
	}

	// Try adding the next item
	if currentWeight+items[nextIndex].weight <= allowedWeight {
		items[nextIndex].isSelected = true
		nextValue := items[nextIndex].value + currentValue
		nextWeight := items[nextIndex].weight + currentWeight
		remainingVal := remainingValue - items[nextIndex].value
		test1Solution, test1Value, test1Calls = doBranchAndBound(items, allowedWeight, nextIndex+1, bestValue, nextValue, nextWeight, remainingVal)
	} else {
		test1Solution, test1Value, test1Calls = nil, 0, 1
	}

	// Try not adding the next item
	// See if there's a chance of improvement without this item's value.
	items[nextIndex].isSelected = false
	remainingVal := remainingValue - items[nextIndex].value
	test2Solution, test2Value, test2Calls = doBranchAndBound(items, allowedWeight, nextIndex+1, bestValue, currentValue, currentWeight, remainingVal)

	// return the better solution
	if test1Value >= test2Value {
		return test1Solution, test1Value, test1Calls + test2Calls + 1
	}
	return test2Solution, test2Value, test1Calls + test2Calls + 1
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
func branchAndBound(items []Item, allowedWeight int) ([]Item, int, int) {
	bestValue := 0
	currentValue := 0
	currentWeight := 0
	remainingValue := sumValues(items, true)
	return doBranchAndBound(items, allowedWeight, 0, bestValue, currentValue, currentWeight, remainingValue)
}

func doBranchAndBound(items []Item, allowedWeight, nextIndex, bestValue, currentValue, currentWeight, remainingValue int) ([]Item, int, int) {
	var test1Solution, test2Solution []Item
	var test1Value, test1Calls, test2Value, test2Calls int

	if nextIndex >= len(items) {
		copyOfItems := copyItems(items)
		return copyOfItems, currentValue, 1
	}

	// we do not have a full assignment.  Can we improve on this solution so it's worth continuing
	if currentValue+remainingValue <= bestValue {
		// No, we can't improve on the best solution found so far
		return nil, currentValue, 1
	}

	// Try adding the next item
	if currentWeight+items[nextIndex].weight <= allowedWeight {
		items[nextIndex].isSelected = true
		nextValue := items[nextIndex].value + currentValue
		nextWeight := items[nextIndex].weight + currentWeight
		remainingVal := remainingValue - items[nextIndex].value
		test1Solution, test1Value, test1Calls = doBranchAndBound(items, allowedWeight, nextIndex+1, bestValue, nextValue, nextWeight, remainingVal)
	} else {
		test1Solution, test1Value, test1Calls = nil, 0, 1
	}

	// Try not adding the next item
	// See if there's a chance of improvement without this item's value.
	items[nextIndex].isSelected = false
	remainingVal := remainingValue - items[nextIndex].value
	test2Solution, test2Value, test2Calls = doBranchAndBound(items, allowedWeight, nextIndex+1, bestValue, currentValue, currentWeight, remainingVal)

	// return the better solution
	if test1Value >= test2Value {
		return test1Solution, test1Value, test1Calls + test2Calls + 1
	}
	return test2Solution, test2Value, test1Calls + test2Calls + 1
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

	// Rod's method
	if numItems > 85 {
		fmt.Println("Too many items for Rod's Method j")
	} else {
		runAlgorithm(RodsTechnique, items, allowedWeight)

	}
	// Branch and bound search
	if numItems > 45 { // Only run branch and bound search if numItems <= 45.
		fmt.Println("Too many items for exhaustive search")
	} else {
		fmt.Println("*** branch and bound Search ***")
		runAlgorithm(branchAndBound, items, allowedWeight)
	}
}

/*
func runAlgorithm(alg func([]Item, int) ([]Item, int, int), items []Item, allowedWeight int) {
*/
