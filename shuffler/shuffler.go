// Packages and Initialization video code.  Now using my shuffle that really does work.  And also now
// includes his weighted shuffler in which the larger the weight, the more likely that it will be near
// the beginning of the list.  Like a roulette wheel with big and small pockets.  The pocket size is
// determined by its weight.

package shuffler

import (
	//	"fmt"
	//	"net/http"
	//	"io/ioutil"
	"math/rand"
	"time"
	//	"sync/atomic"
)

type Shuffleable interface {
	Len() int
	Swap(i, j int)
}

type WeightedShuffleable interface {
	Shuffleable
	Weight(i int) int
}

func init() { // always here, even if I don't define one.  IE, there is a default init() func.
	t := time.Now()
	ns := t.Nanosecond()
	rand.Seed(int64(ns))
}

// Based on my better algorithm, not his inferior algorithm in which the first element was mostly
// returned as the last element.  IE, the shuffling was not very good.
func Shuffle(s Shuffleable) { // the interface is what is passed into the routine
	//									for i := 0; i < s.Len(); i++ {
	for i := s.Len() - 1; i > 0; i-- { // specifically excluding the first element, element[0]
		j := rand.Intn(i) // This returns a random int 0..i, since first element is element[0]
		s.Swap(i, j)
	}
}

func WeightedShuffle(w WeightedShuffleable) {
	totalw := 0
	for i := 0; i < w.Len(); i++ {
		totalw += w.Weight(i)
	}

	// Based on my better algorithm, not his inferior algorithm in which the first element was mostly
	// returned as the last element.  IE, the shuffling was not very good.

	for i := w.Len() - 1; i > 0; i-- { // specifically excluding the first element, element[0]
		pos := rand.Intn(totalw)
		cumw := 0 // cumulative weight
		for j := w.Len() - 1; j >= i; j-- {
			cumw += w.Weight(j)
			if pos >= cumw { // if position is within the bucket to be swapped
				totalw -= w.Weight(j)
				w.Swap(i, j)
				break // done finding one to swap
			}
		}
		j := rand.Intn(i)
		w.Swap(i, j)
	}
}
