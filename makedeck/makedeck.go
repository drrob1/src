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
*/

const lastAltered = "Apr 9, 2022"
const numOfDecks = 1000 // used to be 8.
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
	fmt.Printf("BlackJack Simulation Prgram to write the deck of cards, written in Go.  Last altered %s, compiled by %s \n", lastAltered, runtime.Version())

	flag.BoolVar(&verboseFlag, "v", false, " Verbose mode")
	flag.Parse()

	const deckExtDefaultBinary = ".deck"
	const deckExtDefaultJson = ".json"

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

	shuffleAmount := date.Nanosecond()/1000 + date.Second() + date.Minute() + date.Day() + date.Hour() + date.Year()
	if verboseFlag {
		fmt.Printf(" Will shuffle %d times.\n", shuffleAmount)
	}
	progBar := pb.Default(int64(shuffleAmount))

	shuffleStartTime := time.Now()
	for i := 0; i < shuffleAmount; i++ {
		rand.Shuffle(len(deck), swapFnt)
		progBar.Add(1)
		//rand.Shuffle(len(deck), swapfnt)  I think this is too much.
	}
	timeToShuffle := time.Since(shuffleStartTime) // timeToShuffle is a Duration type, which is an int64 but has methods.
	if verboseFlag {
		fmt.Printf(" It took %s to shuffle %d cards.\n", timeToShuffle.String(), NumOfCards)
	}

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

// ----------------------------------------------------------------------
/*func (l *ListType) SaveBinary(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	encoder := gob.NewEncoder(f)
	err = encoder.Encode(*l)
	return err // I want to make sure that the write operation occurs before the close operation.
}
*/
func pause() {
	fmt.Printf(" hit any key to continue   ")
	var ans string
	fmt.Scanln(&ans)
	fmt.Printf("%s\n", ans)
}
