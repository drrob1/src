// Packages and Initialization video code.  Now with weighted shuffle

package main

import (
	"fmt"
	"time"
	//	"math/rand"
	"src/shuffler"
)

type intSlice []int

func (i_s intSlice) Len() int {
	return len(i_s) // len is the number of elements in the slice
}

func (i_s intSlice) Swap(i, j int) {
	i_s[i], i_s[j] = i_s[j], i_s[i]
}

type strSlice []string

func (s strSlice) Len() int {
	return len(s) // len is number of elements in the slice
}
func (s strSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type weightedString struct {
	weight int
	s      string
}
type wtstrSlice []weightedString

func (w wtstrSlice) Weight(i int) int {
	return w[i].weight
}
func (w wtstrSlice) Len() int {
	return len(w) // len here means # of elements in the slice, not length of its string field
}
func (w wtstrSlice) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}
func main() {
	t := time.Now()
	h, m, s := t.Clock()
	ns := t.Nanosecond()

	fmt.Println("Time from time.Now() is", t)
	fmt.Println(h, ":", m, ":", s, " nanosec ", ns)

	//	rand.Seed(int64(ns));  This has been moved to the init() of shuffler package.

	i_s := intSlice{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	shuffler.Shuffle(i_s)
	fmt.Println(i_s)

	shuffler.Shuffle(i_s)
	fmt.Printf("%d\n", i_s)

	S := strSlice{"the", "quick", "brown", "fox"}
	shuffler.Shuffle(S)
	fmt.Println(S)

	fmt.Println()

	// Now for weighted string stuff, added in the package for this lesson

	wss := wtstrSlice{weightedString{100, "Hello"}, weightedString{200, "World"}, weightedString{10, "Goodbye"}}

	fmt.Println(wss)
	shuffler.WeightedShuffle(wss)
	fmt.Println(wss)
}
