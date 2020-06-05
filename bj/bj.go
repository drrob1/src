package main
/*
  BlackJack Simulator.
  Translated from Modula-2 that I wrote ca 95, and then converted to Windows in 2005.

  The strategy matrix is read in and consists of columns 1 .. 10 (here, 0 .. 9) where ace is first, and last is all 10 value cards.
  Each row is what strategy to follow when your hand totals that row number.  Since I'm dealing w/ integers, I'll discard what is not
  convenient.  IE, I'll not convert to a zero origin array as I think that would be too confusing.  If as I write the code, I change
  my mind, so be it.
 */
import (
	"fmt"
)

const lastAltered = "June 10, 2020"
/*
  REVISION HISTORY
  ======== =======
   4 Jun 20 -- Started to convert the old blackjack Modula-2 code to Go.  This will take a while.

 */

var OptionName []string{"S  ", "H  ", "D  ", "SP ", "SUR"} // Stand, Hit, Double, Split, Surrender

const Ace = iota
const (
	Stand = iota
	Hit
	Double
	Split
	Surrender
)

type OptionRowType [10]int // first element is for the ace, last element is for all 10 value cards (10, jack, queen, king)

var Strategy [22]OptionRowType // Modula-2 ARRAY [5..21] OF OptionRowType.  I'm going to ignore rows that are not used.
var SoftStrategy [12]OptionRowType // Modula-2 ARRAY [2..11] of OptionRowType.  Also going to ignore rows that are not used.
var PairStrategy [11]OptionRowType // Modula-2 ARRAY [1..10] of OptionRowType.  Same about unused rows.
var StrategyErrorFlag bool  // not sure if I'll need this yet.

const numOfDecks = 8
const maxNumOfPlayers = 7
const maxNumOfHands = 1_000_000_000 // 1 million, for now.
const HandsPerPlayer = 7 // I guess this means splitting hands, which can get crazy.
const NumOfCards = 52 * numOfDecks

var resultNames = []string{"  lost", "  pushed", "  won", "  surrend", "  LostDbl", "  WonDbl", "  LostToBJ", "  PushedBJ", "  WonBJ"}
const (
	lost = iota
	pushed
	won
	surrend
	lostdbl
	wondbl
	losttoBJ
	pushedBJ
	wonBJ
	)

type handType struct {
	card1, card2, total int
	doubledflag, surrenderedflag, bustedflag, BJflag, softflag bool
	result int
}

var resplitAcesFlag, lastHandWinLoseFlag, readyToShuffleFlag bool

var player []handType
var dealer handType
var splitsArray []int  // well, slice, actually.  But nevermind this.
var prevResult []int
var numOfPlayers int
var totalWins, totalLosses, totalPushes, totalDblWins, totalDblLosses, totalBJwon, totalBJpushed, totalBJwithDealerAce, totalSplits,
    totalDoubles, totalSurrenders, totalBusts, totalHands int
var score, winsInARow, lossesInARow int
var runs []int

func ReadStrategy() {

}

func WriteStrategy() {

}

func main() {
	fmt.Printf("BlackJack Simulation Prgram, written in Go.  Last altered %s \n", lastAltered)

	deck := make([]int, 0, NumOfCards)

}
