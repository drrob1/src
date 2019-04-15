// interfaces example.  But I then figured out using godoc how to use time to seed rand.

package main

import (
	"fmt"
	//	"net/http"
	//	"io/ioutil"
	"math/rand"
	"time"
	//	"sync/atomic"
)

type shuffler interface {
	Len() int
	Swap(i, j int)
}

func shuffle(s shuffler) { // the interface is what is passed into the routine
	for i := 0; i < s.Len(); i++ {
		j := rand.Intn(s.Len() - i)
		s.Swap(i, j)
	}
}

type intSlice []int

func (i_s intSlice) Len() int {
	return len(i_s)
}

func (i_s intSlice) Swap(i, j int) {
	i_s[i], i_s[j] = i_s[j], i_s[i]
}

type strSlice []string

func (s strSlice) Len() int {
	return len(s)
}

func (s strSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func main() {
	t := time.Now()
	h, m, s := t.Clock()
	ns := t.Nanosecond()

	fmt.Println("Time from time.Now() is", t)
	fmt.Println(h, ":", m, ":", s, " nanosec ", ns)

	rand.Seed(int64(ns))

	i_s := intSlice{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	shuffle(i_s)
	fmt.Println(i_s)

	shuffle(i_s)
	//	fmt.Printf("%q\n",i_s);
	fmt.Printf("%d\n", i_s)

	S := strSlice{"the", "quick", "brown", "fox"}
	shuffle(S)
	fmt.Println(S)

	fmt.Println()
}
