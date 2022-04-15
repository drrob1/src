package main

import (
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	pb "github.com/schollz/progressbar/v3"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

/*
  REVISION HISTORY
  ======== =======
   9 Apr 22 -- While on boat, I'm coding this rtn to make the very big shoe.  I've been thinking about it all week.
  10 Apr 22 -- Now home from boat, and I changed the name from makedeck.go to cardshuffler.go.  Runs of 1000 decks on linux-laptop took 7-10 min each and ~500K iterations.
  11 Apr 22 -- Leox: 1 million estimated at 2 weeks by progressBar, 100K estimated at 32 hrs.  I'll change the shuffling amount.  Current estimate for thelio is ~ 1/2 hr.  I'll wait.
*/

const lastAltered = "Apr 11, 2022"

//const numOfDecks = 100_000 // used to be 8 and was a const.  1 million was estimated at 2 weeks by progressbar.  100K estimated to take 32 hours.  I plan on waiting for it here on leox.
const numOfDecks = 500_000 // On leox, this is estimated to take 14 hrs.
const NumOfCards = 52 * numOfDecks

var deck []int

var clearscreen map[string]func()

var verboseFlag bool

// ------------------------------------------------------- init -----------------------------------
func init() {
	clearscreen = make(map[string]func())
	clearscreen["linux"] = func() { // this is a closure, or an anonymous function
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clearscreen["windows"] = func() { // this is a closure, or an anonymous function
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// ------------------------------------------------------- InitDeck -----------------------------------

func InitDeck() { // Initalize the deck of cards.
	for i := 0; i < 4*numOfDecks; i++ {
		for j := 1; j <= 10; j++ { // There is no card Zero
			deck = append(deck, j)
		}
		deck = append(deck, 10) // Jack
		deck = append(deck, 10) // Queen
		deck = append(deck, 10) // King.
	}
}

// ------------------------------------------------------- main -----------------------------------
// ------------------------------------------------------- main -----------------------------------
func main() {
	fmt.Printf("BlackJack Simulation Program to write the deck of cards, written in Go.  Last altered %s, compiled by %s \n", lastAltered, runtime.Version())

	flag.BoolVar(&verboseFlag, "v", false, " Verbose mode")
	flag.Parse()

	const deckExtDefaultBinary = ".deck"
	const deckExtDefaultJson = ".json"
	var shufflingAmount int

	deck = make([]int, 0, NumOfCards)
	if verboseFlag {
		fmt.Printf(" Deck has %d cards \n", NumOfCards)
	}

	var filenameBinary, filenameJson string

	if flag.NArg() == 0 {
		numofdecksStr := strconv.Itoa(numOfDecks)
		filenameBinary = numofdecksStr + "decks" + deckExtDefaultBinary
		filenameJson = numofdecksStr + "decks" + deckExtDefaultJson
	} else {
		BaseFilename := flag.Arg(0) // will be empty if no name was given.
		filenameBinary = BaseFilename + deckExtDefaultBinary
		filenameJson = BaseFilename + deckExtDefaultJson
	}

	if verboseFlag {
		fmt.Printf(" Filename for the deck is %s and %s\n", filenameBinary, filenameJson)
	}

	// Init the deck and SurrenderStrategyMatrix, and shuffle the deck
	InitDeck()
	//initSurrenderStrategyMatrix() // only used when the surrender option is wanted in the Strategy matrix.

	if verboseFlag {
		fmt.Printf(" Initialized deck.  There are %d:%d cards in the deck.\n", len(deck), NumOfCards)
	}

	date := time.Now()
	rand.Seed(int64(date.Nanosecond()))

	//       need to shuffle here
	swapFnt := func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	}

	if numOfDecks < 10_000 { // just allowing me to change this value and not to have to remember to change the shufflingAmount calculation.
		shufflingAmount = date.Nanosecond() / 1000
	} else {
		shufflingAmount = date.Nanosecond() / 10_000
	}
	shufflingAmount += date.Second() + date.Minute() + date.Day() + date.Hour() + date.Year()
	if verboseFlag {
		fmt.Printf(" Will shuffle %d times.\n", shufflingAmount)
	}
	progBar := pb.Default(int64(shufflingAmount))

	shuffleStartTime := time.Now()
	for i := 0; i < shufflingAmount; i++ {
		rand.Shuffle(len(deck), swapFnt)
		progBar.Add(1)
	}
	timeToShuffle := time.Since(shuffleStartTime) // timeToShuffle is a Duration type, which is an int64 but has methods.
	fmt.Printf(" It took %s to shuffle %d cards.\n", timeToShuffle.String(), NumOfCards)

	t1 := time.Now()
	js, err := json.Marshal(deck)

	if err != nil {
		fmt.Printf(" Error from json marshal is %v.\n", err)
		os.Exit(1)
	}
	err = os.WriteFile(filenameJson, js, 0666)
	if err != nil {
		fmt.Printf(" Error from WriteFile is %v.\n", err)
	}

	f, er := os.Create(filenameBinary)
	if er != nil {
		fmt.Printf(" Error from Create %s is %v.\n", filenameBinary, er)
	}
	defer f.Close()

	encoder := gob.NewEncoder(f)
	err = encoder.Encode(deck)
	if err != nil {
		fmt.Printf(" Error from Encode is %v\n", err)
	}

	fmt.Printf(" Elapsed time for the file writing for both json and binary is %s.\n", time.Since(t1))

} // main
