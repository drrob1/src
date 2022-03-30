package main

/*
  BlackJack Simulator.
  Approximately translated from Modula-2 that I wrote ca 95, and then converted to Windows in 2005.  I changed much of the logic, and decided to use recursion for the
  handling of split hands.  This is my first use of recursion in my own code.

  The strategy matrix is read in and consists of columns 1 .. 10 (here, 0 .. 9) where ace is first, and last is all 10 value cards.
  Each row begins w/ a hand total, and means what strategy to follow when your hand totals that hand total.
  These are the integers from 5 .. 21, then S2 .. S10, meaning soft 2 thru soft 10.  Soft 2 would be 2 aces, but since that's a pair,
  this row is ignored.  Then we have the pairs AA thru 99 and 1010.
  Since I'm dealing w/ integers, I'll ignore row indicies that are not convenient.  IE, ignore Strategy rows < 5, etc.

  The input file now will have a .strat extension, just to be clear.  And the output file will have same basefilename with .results extension.

  I changed the underslying logic a lot from what I wrote 25 years ago in Modula-2.  Now I use recursion to solve the problem of splitting.

  Now, when PairStrategy says to hit, it gets passed to Strategy which has to have an entry for it.  This was a problem for deuces.

  Getting it right when player pulls an Ace is hard.
  I added the clearscreen logic from rpng to help me w/ the debugging screen.

  I have to make sure that double only happens on first 2 cards.

  Split aces cannot double sometimes.  I'll assume that if re-splitting aces is allowed, then they can double also.
  Otherwise, not.

  Surrendering is more complex and has to be re-thought and tested.  I have to make sure that the strategy doesn't surrender after cards have been taken in that round.
  I just added the moreCardsTaken flag, and I'll see if that approach works.  But I just saw that I already have a cannotDouble local flag that may serve the same purpose.
  This may be the reason that Modula-2 got surrender wrong -- all totals of 16 were surrendered, not just when the first 2 cards totaled 16, for example.
  I have to still test dealer standing on soft 17.
  I have to still test the variations on splitting aces, re-splitting aces, doubling split aces
  The 2 flags in the strategy file are dealer17 and resplit.
  Looks like these work as intended.

  I forgot to compute the maxruns, and output the slices of runs.
  And for tomorrow, I'll create and output ratio matricies, where each entry is wins/(wins+losses).  I don't have to also construct loss ratio mactrices.

  Looks like my original classic strategy is almost correct.  But if surrender is allowed, sur1516.strat is optimal.

  And I'll add something just learned w/ the stats output Jun 15, 2020.  This reinforces not to call "even money" when have BJ and dealer shows an Ace.
  Simulator shows 0.075 of getting BJ when dealer ace is showing.  And .046 of BJ pushes.
  Thinking about this some more, I realized that "getting BJ w/ dealer ace showing" includes Bj pushes.  So about 60% of the time I get BJ w/ dealer ace
  showing, I push a BJ.  But BJ pays 3:2.  That makes it ~ 60:40.  So even odds is close enough to being right.  In the future, I'll take even money.

*/
import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"src/tknptr"
	"strconv"
	"strings"
	"time"
	//"getcommandline"  I really haven't needed this for years.
	//"io/ioutil"
)

/*
  REVISION HISTORY
  ======== =======
   4 Jun 20 -- Started to convert the old blackjack Modula-2 code to Go.  This will take a while.
   8 Jun 20 -- First working version in which the logic is correct (I think).  I'm going to test further before
                 coding the collection and display of stats.  Of course, this is all about the stats.
   9 Jun 20 -- I believe the logic is correct.  I'm going to start coding the statistics.
  11 Jun 20 -- Fixing computation of runs if surrender is allowed.  Can't have only 2 states for last hand.
  12 Jun 20 -- Working as needed.  I'll consider this the first fully working version.
  15 Jun 20 -- Now to putput some more stats.
  29 Mar 22 -- Fix it to compile under Go 1.18.
                 I added bytes.Reader instead of bytes.Buffer for reading in the strategy matrix.
                 I added use of a strings.Builder instead of my += construct for building the output line.
  30 Mar 22 -- Will allow comments to start w/ '#' as in bash, and '/' as almost like C-ish.  The change is in readLine.
*/

const lastAltered = "Mar 29, 2022"

var OptionName = []string{"Stnd", "Hit ", "Dbl ", "SP  ", "Sur "} // Stand, Hit, Double, Split, Surrender

const Ace = 1
const ( // strategy codes, originally an enumeration
	Stand = iota
	Hit
	Double
	Split
	Surrender
	ErrorValue
)

const ( // result codes, originally an enumeration
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

type OptionRowType []int // first element is for the ace, last element is for all 10 value cards (10, jack, queen, king)
// by making this a slice, I can append the rows as I read them from the input strategy file.

var StrategyMatrix [22]OptionRowType // Modula-2 ARRAY [5..21] OF OptionRowType.  I'm going to ignore rows that are not used.

// Because of change in logic, SoftStrategyMatrix needs enough rows to handle more cases.  Maybe.  I changed the soft hand logic also.

var SoftStrategyMatrix [22]OptionRowType // Modula-2 ARRAY [2..11] of OptionRowType.  Also going to ignore rows that are not used.

var PairStrategyMatrix [11]OptionRowType // Modula-2 ARRAY [1..10] of OptionRowType.  Same about unused rows.

var SurrenderStrategyMatrix [17]OptionRowType // This can be hard coded because I only consider surrendering 14, 15, 16.

const numOfDecks = 8

const maxNumOfPlayers = 100       // used for the make function on playerHand.
const maxNumOfHands = 100_000_000 // 100 million, for now.  Should be about 20 sec on leox, but the new Ryzen 9 5950X computers are ~half that.
const NumOfCards = 52 * numOfDecks

var resultNames = []string{"lost", "pushed", "won", "surrend", "LostDbl", "WonDbl", "LostToBJ", "PushedBJ", "WonBJ"}

type handType struct {
	card1, card2, total                                                                         int
	notAvirginflag, doubledflag, surrenderedflag, bustedflag, pair, softflag, splitflag, BJflag bool
	result                                                                                      int
}

var displayRound, resplitAcesFlag, dealerHitsSoft17 bool
var playerHand []handType
var hand handType
var dealerHand handType
var numOfPlayers, currentCard int
var deck []int

//var prevResult []int
var runsWon, runsLost []int
var lastHandWinLose int = ErrorValue // this cannot be a bool to correctly count surrender.  Not having it zero means that the first hand is counted correctly, also.
var currentRunWon, currentRunLost int
var totalWins, totalLosses, totalPushes, totalDblWins, totalDblLosses, totalBJwon, totalBJpushed, totalBJwithDealerAce, totalSplits,
	totalDoubles, totalSurrenders, totalBusts, totalHands int

//var winsInARow, lossesInARow int
var score float64
var clearscreen map[string]func()

// Stats arrays
type statsRowType [11]int // Here, unlike the strategy matrices, I'll use cards as column numbers.  Using card-1 is confusing to look at.

// I will use the first 2 card totals as the row of the matrix, and the column will be dealer.card1, where Ace is 1.  This will have
// empty rows.  But this is too small an amount of waste to bother about.
var WonStats, LostStats, DoubleWonStats, DoubleLostStats, SoftWonStats, SoftLostStats, SoftDoubleWonStats, SoftDoubleLostStats [22]statsRowType

type ratioRowType [11]float64

var ratioWon, ratioDoubleWon, ratioSoftWon, ratioSoftDoubleWon [22]ratioRowType

var OutputFilename string
var OutputHandle *os.File
var bufOutputFileWriter *bufio.Writer
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

// ------------------------------------------------------- InitSurrenderStrategy -----------------------------------
func initSurrenderStrategyMatrix() {
	var row14 = OptionRowType{Hit, Stand, Stand, Stand, Stand, Stand, Hit, Hit, Hit, Hit}
	var row15 = OptionRowType{Hit, Stand, Stand, Stand, Stand, Stand, Hit, Hit, Hit, Hit}
	var row16 = OptionRowType{Hit, Stand, Stand, Stand, Stand, Stand, Hit, Hit, Hit, Hit}

	SurrenderStrategyMatrix[14] = row14
	SurrenderStrategyMatrix[15] = row15
	SurrenderStrategyMatrix[16] = row16
}

// ------------------------------------------------------- GetOption -----------------------------------

func GetOption(tkn tknptr.TokenType) int {
	if tkn.Str == "S" {
		return Stand
	} else if tkn.Str == "H" {
		return Hit
	} else if tkn.Str == "D" {
		return Double
	} else if tkn.Str == "SP" {
		return Split
	} else if tkn.Str == "SUR" {
		return Surrender
	} else {
		return ErrorValue
	}
}

// ------------------------------------------------------- ReadStrategyMatrix -----------------------------------

//func ReadStrategyMatrix(buf *bytes.Buffer) {

func ReadStrategyMatrix(buf *bytes.Reader) { // the StrategyMatrix is global.
	if verboseFlag {
		fmt.Printf(" Entering ReadStrategyMatrix\n")
	}
	for {
		//rowbuf, err := buf.ReadString('\n')
		rowbuf, err := readLine(buf)
		if verboseFlag {
			fmt.Printf(" read a line using readLine.  rowbuf=%q, len(rowbuf)=%d, err= %v\n", rowbuf, len(rowbuf), err)
			pause()
		}
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(" Error from readLine is", err)
			os.Exit(1)
		}
		//rowbuf = strings.TrimSpace(rowbuf)  redundant as this is done by readLine now.
		if len(rowbuf) == 0 { // ignore blank lines
			continue
		}
		tknbuf := tknptr.NewToken(rowbuf)
		rowID, EOL := tknbuf.GetToken(true) // force upper case token.
		if EOL {
			return
		}
		row := make([]int, 0, 10) // a single StrategyMatrix row.
		for {
			token, eol := tknbuf.GetToken(true) // force upper case token
			if eol {
				break
			}
			if EOL || token.State != tknptr.ALLELSE {
				return
			}
			i := GetOption(token)
			if i == ErrorValue { // allow for comments after the StrategyMatrix codes on same line.
				return
			}
			row = append(row, i)
		}

		if rowID.Isum >= 4 && rowID.Isum <= 21 { // assign StrategyMatrix codes to proper row in StrategyMatrix Decision Matrix
			StrategyMatrix[rowID.Isum] = row
		} else if rowID.State == tknptr.DGT {
			switch rowID.Isum {
			case 22:
				PairStrategyMatrix[2] = row
			case 33:
				PairStrategyMatrix[3] = row
			case 44:
				PairStrategyMatrix[4] = row
			case 55:
				PairStrategyMatrix[5] = row
			case 66:
				PairStrategyMatrix[6] = row
			case 77:
				PairStrategyMatrix[7] = row
			case 88:
				PairStrategyMatrix[8] = row
			case 99:
				PairStrategyMatrix[9] = row
			case 1010:
				PairStrategyMatrix[10] = row
			default:
				fmt.Println(" Invalid Pair Row value:", rowID) // rowID is a struct, so all of it will be output.
				fmt.Print(" continue? y/n ")
				var ans string
				n, err := fmt.Scanln(&ans)
				if err != nil || n == 0 {
					ans = "n"
				}
				ans = strings.ToLower(ans)
				if ans != "y" {
					os.Exit(1)
				}
			} // end switch-case for pairs.

		} else if rowID.State == tknptr.ALLELSE {
			if rowID.Str == "AA" {
				PairStrategyMatrix[1] = row
			} else if rowID.Str[0] == 'S' { // soft hand
				s := rowID.Str[1:] // chop off the "S" and convert rest to int
				i, err := strconv.Atoi(s)
				if err != nil {
					fmt.Println(" Error in reading soft hand designation beginning w/ S:", err)
					fmt.Print(" continue? y/n ")
					var ans string
					_, _ = fmt.Scanln(&ans)
					ans = strings.ToLower(ans)
					if ans != "y" {
						os.Exit(1)
					}
				}
				SoftStrategyMatrix[i] = row
			} else if rowID.Str[0] == 'A' { // First card is an Ace, ie a soft hand, but notation is different.
				A := rowID.Str[1:]        // chop off the "A" and convert rest to int
				i, err := strconv.Atoi(A) // i is off by one b/o value of Ace is 11, not 10.
				if err != nil {
					fmt.Println(" Error in reading soft hand designation beginning w/ A:", err)
					fmt.Print(" continue? y/n ")
					ans := ""
					_, _ = fmt.Scanln(&ans)
					ans = strings.ToLower(ans)
					if ans != "y" {
						os.Exit(1)
					}
				}
				SoftStrategyMatrix[i+1] = row
			} else if rowID.Str == "DEALER17" {
				dealerHitsSoft17 = true
			} else if rowID.Str == "RESPLIT" {
				resplitAcesFlag = true

			} else {
				fmt.Println(" Invalid Row value:", rowID) // rowID is a struct, so all of it will be output.

				fmt.Print(" continue? y/n ")
				ans := ""
				_, _ = fmt.Scanln(&ans)
				ans = strings.ToLower(ans)
				if ans != "y" {
					os.Exit(1)
				}
			}
		}

	}
	if verboseFlag {
		fmt.Printf(" Leaving ReadStrategyMatrix.\n\n")
	}
} // ReadStrategyMatrix

// ------------------------------------------------------- WriteStrategy -----------------------------------

func WriteStrategyMatrix(filehandle *bufio.Writer) {
	var sb strings.Builder
	filehandle.WriteString(" Regular Strategy Matrix: \n")

	// First write out regular StrategyMatrix
	for i, row := range StrategyMatrix {
		if i < 5 { // ignore rows < 5, as these are special cases and are in the other matrixes
			continue
		}
		outputline := fmt.Sprintf(" %2d: ", i)
		sb.WriteString(outputline)
		for _, j := range row {
			s := fmt.Sprintf("%s  ", OptionName[j])
			sb.WriteString(s)
			//outputline += s
		}
		sb.WriteRune('\n')
		//filehandle.WriteString(outputline)
		filehandle.WriteString(sb.String())
		//filehandle.WriteRune('\n')
		sb.Reset()
	}

	// Now write out Soft StrategyMatrix
	filehandle.WriteString(" \n \n Soft Strategy Matrix, IE, Strategy for soft hands: \n")
	for i, row := range SoftStrategyMatrix {
		if i < 3 { // ignore row 0, 1 and 2, as these are not a valid blackjack hand or are in PairStrategyMatrix
			continue
		}
		outputline := fmt.Sprintf(" s%2d: ", i)
		sb.Reset()
		sb.WriteString(outputline)
		for _, j := range row {
			s := fmt.Sprintf("%s  ", OptionName[j])
			sb.WriteString(s)
			//outputline += s
		}
		sb.WriteRune('\n')
		//filehandle.WriteString(outputline)
		filehandle.WriteString(sb.String())
		//filehandle.WriteRune('\n')
	}

	// Now write out Pair StrategyMatrix
	filehandle.WriteString(" \n \n Pair Strategy Matrix: \n")
	for i, row := range PairStrategyMatrix {
		if i < 1 { // ignore row 0 as it is not a valid blackjack hand
			continue
		}
		outputline := fmt.Sprintf(" %2d-%2d: ", i, i)
		sb.Reset()
		sb.WriteString(outputline)
		for _, j := range row {
			s := fmt.Sprintf("%s  ", OptionName[j])
			sb.WriteString(s)
			//outputline += s
		}
		sb.WriteRune('\n')
		//filehandle.WriteString(outputline)
		filehandle.WriteString(sb.String())
		//filehandle.WriteRune('\n')
	}

	filehandle.WriteRune('\n')
	filehandle.WriteRune('\n')
	filehandle.WriteRune('\n')
} // WriteStrategyMatrix

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

// ------------------------------------------------------- doTheShuffle -----------------------------------
func doTheShuffle() {
	currentCard = 0
	swapfnt := func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	}
	rand.Shuffle(len(deck), swapfnt)
}

// ------------------------------------------------------- getCard -----------------------------------
func getCard() int {
	currentCard++ // This will ignore the first card, in position zero.
	return deck[currentCard]
}

// ------------------------------------------------------- hitDealer -----------------------------------
// This plays until stand or bust.
func hitDealer() {
	if displayRound {
		fmt.Printf(" Entering DealerHit.  Hand is: %d %d, total=%d \n", dealerHand.card1, dealerHand.card2, dealerHand.total)
	}
	if dealerHand.BJflag {
		return
	}
	for dealerHand.total < 18 { // always exit if >= 18.  Loop if < 18.
		if displayRound {
			fmt.Printf("\n Top of for Dealerhand total.  Hand is: %d %d, total=%d, soft=%t \n",
				dealerHand.card1, dealerHand.card2, dealerHand.total, dealerHand.softflag)
		}
		if dealerHand.softflag && dealerHand.total <= 11 && dealerHitsSoft17 {
			if dealerHand.total > 7 && dealerHand.total <= 11 {
				dealerHand.total += 10
			}
			if dealerHand.total <= 17 {
				if displayRound {
					fmt.Printf(" Dealer soft17 hit loop.  Total=%d \n", dealerHand.total)
				}
				dealerHand.total += getCard()
				if dealerHand.total > 21 {
					dealerHand.bustedflag = true
					return
				} else if dealerHand.total > 7 && dealerHand.total <= 11 {
					dealerHand.total += 10
					return
				}
			} // until busted or stand
		} else if dealerHand.softflag && dealerHand.total <= 11 { // he must stand on a soft 17.
			if dealerHand.total >= 7 && dealerHand.total <= 11 {
				dealerHand.total += 10
			}
			if dealerHand.total < 17 { // until busted or stand
				if displayRound {
					fmt.Printf(" Dealer soft no hit soft17 loop.  Total=%d \n", dealerHand.total)
				}
				dealerHand.total += getCard()
				if dealerHand.total > 21 {
					dealerHand.bustedflag = true
					return
				} else if dealerHand.total >= 7 && dealerHand.total <= 11 {
					dealerHand.total += 10
					return
				}
			} // until busted or stand
		} else { // not soft, or Ace cannot be 11, only can be 1.
			if dealerHand.total < 17 {
				if displayRound {
					fmt.Printf(" Dealer hard hand hit loop.  Total=%d \n", dealerHand.total)
				}
				newcard := getCard()
				if newcard == 1 {
					dealerHand.softflag = true
				}
				dealerHand.total += newcard
				if dealerHand.total > 21 {
					dealerHand.bustedflag = true
					return
				}
			} // until busted or stand
		} // if soft hand or not.
		if dealerHand.softflag && !dealerHitsSoft17 && dealerHand.total >= 17 { // this could probably be == 17 and still work.
			return
		} else if dealerHand.total >= 17 {
			return
		}
	} // for loop, which allows hands to jump for the hard to soft category.
	return
} // hitDealer

// ------------------------------------------------------- hitMePlayer -----------------------------------
// This only takes one card for the playerHand, but for the dealer it plays until stand or bust.
// Note that blackjack has already been checked for, in playAhand().
func hitMePlayer(i int) {
	cannotDouble := false
	if playerHand[i].splitflag && playerHand[i].softflag && !resplitAcesFlag { // assuming that cannot double a soft hand after splitting aces if can't resplit aces.
		cannotDouble = true
	}

MainLoop:
	for { // until stand or bust.  Recall that player hands that need to check the strategy matrices for each iteration, unlike the dealer.
		if displayRound {
			fmt.Printf(" Top of hitMePlayer: playerHand[%d]: card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, split=%t, BJ=%t \n\n",
				i, playerHand[i].card1, playerHand[i].card2, playerHand[i].total, playerHand[i].doubledflag, playerHand[i].surrenderedflag,
				playerHand[i].bustedflag, playerHand[i].pair, playerHand[i].softflag, playerHand[i].splitflag, playerHand[i].BJflag)
		}

		if playerHand[i].softflag && playerHand[i].total <= 11 { // if total > 11, Ace can only be 1 and not a 10.
			strategy := SoftStrategyMatrix[playerHand[i].total][dealerHand.card1-1]
			if strategy == Double && cannotDouble { // can only double on first time thru, else have more than 2 cards.
				strategy = Hit
			} else {
				cannotDouble = true // so next time thru the loop cannot double
			}
			if displayRound {
				fmt.Printf(" SoftStrategy[%d][%d] = %s \n\n", playerHand[i].total, dealerHand.card1-1, OptionName[strategy])
			}
			switch strategy { //SoftStrategyMatrix
			case Stand: //SoftStrategyMatrix
				if playerHand[i].total <= 11 { // here is the logic of a soft hand
					playerHand[i].total += 10
				}
				break MainLoop

			case Hit: //SoftStrategyMatrix
				card := getCard()
				playerHand[i].total += card
				playerHand[i].notAvirginflag = true
				if playerHand[i].total > 21 { // if you hit and bust a soft hand, then the Ace must be 1
					playerHand[i].bustedflag = true
					break MainLoop
				}
				cannotDouble = true

			case Double: // here this means hit only once.  SoftStrategyMatrix
				if !playerHand[i].doubledflag {
					playerHand[i].total += getCard()
					if playerHand[i].total > 21 { // since Ace is initially treated as a 1, this should never bust.  Unless doubling 12+
						playerHand[i].bustedflag = true
					} else if playerHand[i].total <= 11 { // soft hand effect.
						playerHand[i].total += 10
					}
				}
				playerHand[i].doubledflag = true
				playerHand[i].notAvirginflag = true
				break MainLoop

			case Surrender: //SoftStrategyMatrix
				fmt.Printf(" in hitMePlayer for soft hands and got a surrender option.  I=%d, hand=%v \n", i, playerHand[i])
				return

			case ErrorValue: //SoftStrategyMatrix
				fmt.Printf(" in hitMePlayer and got errorValue.  I is %d, and hand is %v \n", i, playerHand[i])
				return
			} // switch-case
		} else { // not a soft hand where Ace can represent 11.
			strategy := StrategyMatrix[playerHand[i].total][dealerHand.card1-1]
			if strategy == Double && cannotDouble { // can only double on first time thru, else have more than 2 cards.
				strategy = Hit
			} else {
				cannotDouble = true // so next time thru the loop cannot double
			}
			if strategy == Surrender && playerHand[i].notAvirginflag {
				strategy = SurrenderStrategyMatrix[playerHand[i].total][dealerHand.card1-1]
			}
			if displayRound {
				fmt.Printf(" StrategyMatrix[%d][%d] = %s \n\n", playerHand[i].total, dealerHand.card1-1, OptionName[strategy])
			}
			switch strategy { // StrategyMatrix
			case Stand: // StrategyMatrix
				if displayRound {
					fmt.Printf(" Hard HitMe - stand.  total= %d \n", playerHand[i].total)
				}
				return

			case Hit: // Hard StrategyMatrix
				if displayRound {
					fmt.Printf(" Hard HitMe - hit.  total= %d \n", playerHand[i].total)
				}
				newcard := getCard()
				playerHand[i].total += newcard
				if newcard == 1 { // if pulled an Ace, next time around this is considered a soft hand.
					playerHand[i].softflag = true
				}
				playerHand[i].notAvirginflag = true
				if playerHand[i].total > 21 {
					playerHand[i].bustedflag = true
					return
				} else if playerHand[i].softflag && playerHand[i].total >= 9 && playerHand[i].total <= 11 {
					playerHand[i].total += 10
				}

			case Double: // Hard hit only once.  Hard StrategyMatrix
				if displayRound {
					fmt.Printf(" Hard HitMe - double.  total= %d \n", playerHand[i].total)
				}
				if !playerHand[i].doubledflag { // must only draw 1 card.
					newcard := getCard()
					playerHand[i].total += newcard
					playerHand[i].notAvirginflag = true
					if playerHand[i].total > 21 {
						playerHand[i].bustedflag = true
						return
					} else if newcard == 1 && playerHand[i].total <= 11 { // make sure total is correct when the one allowed card is an Ace.
						playerHand[i].total += 10
						playerHand[i].softflag = true
					}
				}
				playerHand[i].doubledflag = true
				return

			case Surrender: // Hard StrategyMatrix
				if displayRound {
					fmt.Printf(" Hard HitMe - surrender.  total= %d \n", playerHand[i].total)
				}
				playerHand[i].surrenderedflag = true
				playerHand[i].notAvirginflag = true
				playerHand[i].result = surrend
				return

			case ErrorValue: // Hard StrategyMatrix
				fmt.Printf(" in hard hitMe and got errorValue.  I is %d, and hand is %v \n", i, playerHand[i])
				return
			} // switch case on Strategy.
		} // if soft or not.
	} // until finished taking cards.  This is either standing pat or busted.  There are no other options.
	if playerHand[i].total <= 11 && playerHand[i].softflag {
		dealerHand.total += 10
	}
} // hitMePlayer

// ------------------------------------------------------- dealCards -----------------------------------
func dealCards() {
	for i := range playerHand {
		playerHand[i] = handType{} // init the new hand.
		playerHand[i].card1 = getCard()
		playerHand[i].softflag = (playerHand[i].card1 == Ace)
	}
	dealerHand = handType{} // init the new dealer's hand, also.
	dealerHand.card1 = getCard()
	dealerHand.softflag = (dealerHand.card1 == Ace)
	for i := range playerHand {
		playerHand[i].card2 = getCard()
		playerHand[i].total = playerHand[i].card1 + playerHand[i].card2
		playerHand[i].pair = (playerHand[i].card1 == playerHand[i].card2)
		playerHand[i].softflag = (playerHand[i].softflag || (playerHand[i].card2) == Ace)
		playerHand[i].BJflag = (playerHand[i].total == 11 && playerHand[i].softflag)
	}
	dealerHand.card2 = getCard()
	dealerHand.total = dealerHand.card1 + dealerHand.card2
	dealerHand.pair = (dealerHand.card1 == dealerHand.card2)
	dealerHand.softflag = (dealerHand.softflag || (dealerHand.card2 == Ace))
	dealerHand.BJflag = (dealerHand.total == 11 && dealerHand.softflag)
} // dealCards

// ------------------------------------------------------- splitHand -----------------------------------
func splitHand(i int) {
	if displayRound {
		fmt.Printf("\n Entering splitHand.  card1=%d, card2=%d \n", playerHand[i].card1, playerHand[i].card2)
	}
	hnd := handType{}
	hnd.card1 = playerHand[i].card2
	hnd.card2 = getCard()
	hnd.pair = (hnd.card2 == hnd.card1)
	hnd.softflag = (hnd.card1 == Ace) || (hnd.card2 == Ace)
	hnd.total = hnd.card1 + hnd.card2
	hnd.splitflag = true

	playerHand[i].card2 = getCard()
	if playerHand[i].card1 == playerHand[i].card2 {
		playerHand[i].pair = true
	} else {
		playerHand[i].pair = false
	}
	playerHand[i].total = playerHand[i].card1 + playerHand[i].card2
	playerHand[i].softflag = playerHand[i].card1 == 1 || playerHand[i].card2 == 1
	playerHand[i].splitflag = true

	playerHand = append(playerHand, hnd) // can't get blackjack after a split.
	totalSplits++

	if displayRound {
		fmt.Printf(" splitHand: 1st hand.card1=%d, 1st hand.card2=%d, 1st hand.total=%d; split-off hnd.card1=%d, hnd.card2=%d, hnd.total=%d \n\n ",
			playerHand[i].card1, playerHand[i].card2, playerHand[i].total, hnd.card1, hnd.card2, hnd.total)
		fmt.Println(" Exiting splitHand: playerHand slice =", playerHand)
	}
} // splitHand

// ------------------------------------------------------- playAhand -----------------------------------
func playAhand(i int) {
	if displayRound {
		fmt.Printf(" Top of playAhand: playerHand[%d]: card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, BJ=%t \n",
			i, playerHand[i].card1, playerHand[i].card2, playerHand[i].total, playerHand[i].doubledflag, playerHand[i].surrenderedflag,
			playerHand[i].bustedflag, playerHand[i].pair, playerHand[i].softflag, playerHand[i].BJflag)
		fmt.Printf(" Top of playAhand: dealerHand is card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, BJ=%t \n",
			dealerHand.card1, dealerHand.card2, dealerHand.total, dealerHand.doubledflag, dealerHand.surrenderedflag,
			dealerHand.bustedflag, dealerHand.pair, dealerHand.softflag, dealerHand.BJflag)
		fmt.Println()
	}

	if playerHand[i].BJflag && !dealerHand.BJflag {
		playerHand[i].result = wonBJ
		return
	} else if playerHand[i].BJflag && dealerHand.BJflag {
		playerHand[i].result = pushedBJ
		return
	} else if dealerHand.BJflag {
		playerHand[i].result = losttoBJ
		return
	} else if playerHand[i].pair {
		if playerHand[i].softflag { // this must be the AA, or a pair of aces hand.  This is a special case.
			if displayRound {
				fmt.Println(" About to consider splitting aces.  Have not yet checked to see about resplitting aces.")
			}
			if playerHand[i].splitflag && !resplitAcesFlag { // not allowed to resplit aces (or double the hand).
				return
			}
			splitHand(i)
			playAhand(i) // I'm trying out recursion to solve this problem.  The split-off additional hand will be handled by playAllHands.
			return
		} else {
			strategy := PairStrategyMatrix[playerHand[i].card1][dealerHand.card1-1]
			if displayRound {
				fmt.Printf("playAhand for hand=%d, PairStrategyMatrix[%d][%d] = %s \n\n", i, playerHand[i].total, dealerHand.card1-1, OptionName[strategy])
			}

			switch strategy {
			case Stand:
				return
			case Hit:
				hitMePlayer(i)
			case Double:
				// playerHand[i].doubledflag = true  This prevents taking a card in HitMePlayer.  I'll let StrategyMatrix for 10 take over here.
				// No other pair doubles.
				hitMePlayer(i)
				return // double takes 1 card at most.
			case Split:
				splitHand(i)
				playAhand(i) // recursion.  First time I'm using it, but this problem lends itself to recursion as a solution.
				return
			case Surrender:
				playerHand[i].surrenderedflag = true
				return
			case ErrorValue:
				fmt.Printf(" PairStrategyMatrix returned ErrorValue.  i=%d, hand= %v \n", i, playerHand[i])
				return
			} // switch-case
		} // if hand is soft
	} else { // a regular hand that is not a blackjack or a pair, but could be soft.
		hitMePlayer(i)
	} // if playerhand is blackjack, etc
} // playAhand

// ------------------------------------------------------- playAllHands -----------------------------------
func playAllHands() {
	dealCards() // dealCards must be outside the loop of playerHands, else new cards are dealt for each hand.  Oops.

	for i := 0; i < len(playerHand); i++ { // can't range over hands because splits add to the hands slice.
		playAhand(i)
	}
	hitDealer() // play the dealer's hand.  i is ignored so I'm just using the zero as a filler.
} // playAllHands

// ------------------------------------------------------- showDown -----------------------------------
func showDown() {
	// Here is where I will check the player[i] result field, and splits result field, if there are any splits in this round.
	for i := range playerHand {
		totalHands++
		switch playerHand[i].result {
		case wonBJ:
			totalBJwon++
			if dealerHand.card1 == Ace {
				totalBJwithDealerAce++
			}
		case pushedBJ:
			totalBJpushed++
			if dealerHand.card1 == Ace {
				totalBJwithDealerAce++
			}
		case losttoBJ:
			totalLosses++
		case surrend:
			totalSurrenders++
		default:
			if playerHand[i].bustedflag {
				totalBusts++
				if playerHand[i].doubledflag {
					playerHand[i].result = lostdbl
					totalDblLosses++
				} else {
					playerHand[i].result = lost
					totalLosses++
				}
			} else if dealerHand.bustedflag {
				if playerHand[i].doubledflag {
					playerHand[i].result = wondbl
					totalDblWins++
					totalDoubles++
				} else {
					playerHand[i].result = won
					totalWins++
				}
			} else if playerHand[i].total > dealerHand.total {
				if playerHand[i].doubledflag {
					playerHand[i].result = wondbl
					totalDblWins++
					totalDoubles++
				} else {
					playerHand[i].result = won
					totalWins++
				}
			} else if playerHand[i].total == dealerHand.total {
				playerHand[i].result = pushed
				totalPushes++
			} else if playerHand[i].total < dealerHand.total {
				if playerHand[i].doubledflag {
					playerHand[i].result = lostdbl
					totalDblLosses++
					totalDoubles++
				} else {
					playerHand[i].result = lost
					totalLosses++
				}
			}
		} // seitch-case

	} // for range over all hands, incl'g split hands.
} // showDown

// ------------------------------------------------------- IncrementStats -----------------------------------
// type statsRowType [11]int // Here, unlike the strategy matrices, I'll use cards as column numbers.  Using card-1 is confusing to look at.
// var WonStats, LostStats, DoubleWonStats, DoubleLostStats, SoftWonStats, SoftLostStats, SoftDoubleWonStats, SoftDoubleLostStats [22]statsRowType
// var lastHandWinLoseFlag bool
// var winsInARow, lossesInARow int
// And to compute and store the runs of wins and losses.  Ignore pushes or surrenders.  BJ is a win, double is a single win or loss for sake of runs since
// it's still just 1 hand.  I think I'll include each and every split hand it this as a separate hand.
// The local SoftFlag must be used because the playerHand[i].SoftFlag is set if ANY subsequent cards are an Ace, so it can properly handle soft hands.
func incrementStats() {
	for i := range playerHand { // range over all hannds, including split hands.
		initialPlayerTotal := playerHand[i].card1 + playerHand[i].card2
		FirstDealerCard := dealerHand.card1
		SoftFlag := playerHand[i].card1 == Ace || playerHand[i].card2 == Ace
		switch playerHand[i].result {
		case lost:
			if SoftFlag {
				SoftLostStats[initialPlayerTotal][FirstDealerCard]++
			} else {
				LostStats[initialPlayerTotal][FirstDealerCard]++
			}

			if lastHandWinLose == won {
				lastHandWinLose = lost
				if currentRunWon > 10 {
					runsWon = append(runsWon, currentRunWon)
				}
				currentRunWon = 0
			} else if lastHandWinLose == lost {
				currentRunLost++
			} else { // either pushed or ErrorValue
				lastHandWinLose = lost
			}

		case pushed:
			lastHandWinLose = pushed

		case won:
			if SoftFlag {
				SoftWonStats[initialPlayerTotal][FirstDealerCard]++
				// Don't understand how I can have a soft win in row 11.  Now I do.  Splits don't count as blackjack and are a soft 11.  So it's right after all.
				if displayRound {
					if initialPlayerTotal == 11 {
						fmt.Printf(" SoftWonStats incremented.  playerHand=%v \n", playerHand[i])
						fmt.Print(" hit <enter> to continue  ")
						ans := ""
						fmt.Scanln(&ans)
						if ans == "q" {
							os.Exit(1)
						}
					}
				}

			} else {
				WonStats[initialPlayerTotal][FirstDealerCard]++
			}

			if lastHandWinLose == won {
				currentRunWon++
			} else if lastHandWinLose == lost {
				lastHandWinLose = won
				if currentRunLost > 10 {
					runsLost = append(runsLost, currentRunLost)
				}
				currentRunLost = 0
			} else { // either pushed or ErrorValue
				lastHandWinLose = won
			}

		case surrend:
			lastHandWinLose = surrend // not keeping a run of surrenders.

		case lostdbl:
			if SoftFlag {
				SoftDoubleLostStats[initialPlayerTotal][FirstDealerCard]++
			} else {
				DoubleLostStats[initialPlayerTotal][FirstDealerCard]++
			}

			if lastHandWinLose == won {
				lastHandWinLose = lost
				if currentRunWon > 10 {
					runsWon = append(runsWon, currentRunWon)
				}
				currentRunWon = 0
			} else if lastHandWinLose == lost {
				currentRunLost++
			} else { // pushed or ErrorValue
				lastHandWinLose = lost
			}

		case wondbl:
			if SoftFlag {
				SoftDoubleWonStats[initialPlayerTotal][FirstDealerCard]++

			} else {
				DoubleWonStats[initialPlayerTotal][FirstDealerCard]++
			}

			if lastHandWinLose == won {
				currentRunWon++
			} else if lastHandWinLose == lost {
				lastHandWinLose = won
				if currentRunLost > 10 {
					runsLost = append(runsLost, currentRunLost)
				}
				currentRunLost = 0
			} else { // either pushed or ErrorValue
				lastHandWinLose = won
			}

		case losttoBJ:
			if lastHandWinLose == won {
				lastHandWinLose = lost
				if currentRunWon > 10 {
					runsWon = append(runsWon, currentRunWon)
				}
				currentRunWon = 0
			} else if lastHandWinLose == lost {
				currentRunLost++
			} else { // either pushed or ErrorValue
				lastHandWinLose = lost
			}

		case pushedBJ:
			lastHandWinLose = pushed

		case wonBJ:
			if lastHandWinLose == won {
				currentRunWon++
			} else if lastHandWinLose == lost {
				lastHandWinLose = won
				if currentRunLost > 4 {
					runsLost = append(runsLost, currentRunLost)
				}
				currentRunLost = 0
			} else { // either pushed or ErrorValue
				lastHandWinLose = won
			}

		} // end switch-case
	} // end for range all hands.

} // end incrementStats

// ------------------------------------------------------- wrStatsToFile -----------------------------------
func wrStatsToFile() {
	// declared above
	// The row corresponding to a hand total of 21 in first 2 cards are all zeros, as that's blackjack and handled separately.
	//type statsRowType [11]int // Here, unlike the strategy matrices, I'll use cards as column numbers without subtracting 1.
	//var WonStats, LostStats, DoubleWonStats, DoubleLostStats, SoftWonStats, SoftLostStats, SoftDoubleWonStats, SoftDoubleLostStats [22]statsRowType

	var err error

	bufOutputFileWriter.WriteString(" Won Stats Array \n          A       2       3       4       5       6       7       8       9      10 \n")
	bufOutputFileWriter.WriteString("-------------------------------------------------------------------------------------------------------\n")
	for i := range WonStats {
		if i < 2 || i > 20 {
			continue
		}

		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range WonStats[i] {
			if j == 0 {
				continue
			}
			rowString := ""
			if stats == 0 {
				rowString = "      - "
			} else {
				rowString = fmt.Sprintf(" %7d", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a WonStats row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Lost Stats Array \n")
	bufOutputFileWriter.WriteString("          A       2       3       4       5       6       7       8       9      10 \n")
	bufOutputFileWriter.WriteString("------------------------------------------------------------------------------------\n")
	for i := range LostStats {
		if i < 2 || i > 20 {
			continue
		}

		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range LostStats[i] {
			if j == 0 || j == 21 {
				continue
			}
			rowString := ""
			if stats == 0 {
				rowString = "      - "
			} else {
				rowString = fmt.Sprintf(" %7d", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a LostStats row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Double Won Stats Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range DoubleWonStats {
		if i == 0 {
			continue
		}

		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range DoubleWonStats[i] {
			if j == 0 || j == 21 {
				continue
			}
			rowString := ""
			if stats == 0 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6d", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a DoubleWonStats row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Double Lost Stats Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range DoubleLostStats {
		if i < 2 {
			continue
		}
		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j := 1; j < len(DoubleLostStats[i])-1; j++ { // don't want to display row 21, which is all zeros
			rowString := ""
			if DoubleLostStats[i][j] == 0 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6d", DoubleLostStats[i][j]) // just to show this works
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a DoubleLostStats row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Soft Won Stats Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range SoftWonStats {
		if i < 2 {
			continue
		}
		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range SoftWonStats[i] {
			if j == 0 || j == 21 {
				continue
			}
			rowString := ""
			if stats == 0 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6d", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a SoftWonStats row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Soft Lost Stats Array \n")
	bufOutputFileWriter.WriteString("        A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("-------------------------------------------------------------------------\n")
	for i := range SoftLostStats {
		if i < 2 {
			continue
		}

		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range SoftLostStats[i] {
			if j == 0 || j == 21 {
				continue
			}
			rowString := ""
			if stats == 0 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6d", stats)
			}

			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a SoftLostStats row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Soft Double Won Stats Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range SoftDoubleWonStats {
		if i < 2 {
			continue
		}

		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range SoftDoubleWonStats[i] {
			if j == 0 {
				continue
			}
			rowString := ""
			if stats == 0 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6d", stats)
			}

			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a SoftDoubleWonStats row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Soft Double Lost Stats Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range SoftLostStats {
		if i < 2 {
			continue
		}
		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range SoftDoubleLostStats[i] {
			if j == 0 || j == 21 {
				continue
			}
			rowString := ""
			if stats == 0 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6d", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a SoftDoubleLostStats row.  Error is", err)
		}
	}

	// Compute Ratio Matricies
	// type ratioRowType [11]float64
	// var ratioWon, ratioDoubleWon, ratioSoftWon, ratioSoftDoubleWon [22]ratioRowType
	// And remember, don't divide by zero
	for i := 2; i < len(ratioWon); i++ {
		for j := 1; j < 11; j++ {
			denom := WonStats[i][j] + LostStats[i][j]
			if denom == 0 {
				ratioWon[i][j] = 0
				continue
			}
			ratio := float64(WonStats[i][j]) / float64(denom)
			ratioWon[i][j] = ratio
		}
	}

	for i := 2; i < len(ratioDoubleWon); i++ {
		for j := 1; j < 11; j++ {
			denom := DoubleWonStats[i][j] + DoubleLostStats[i][j]
			if denom == 0 {
				ratioDoubleWon[i][j] = 0
				continue
			}
			ratio := float64(DoubleWonStats[i][j]) / float64(denom)
			ratioDoubleWon[i][j] = ratio
		}
	}

	for i := 2; i < len(ratioSoftWon); i++ {
		for j := 1; j < 11; j++ {
			denom := SoftWonStats[i][j] + SoftLostStats[i][j]
			if denom == 0 {
				ratioSoftWon[i][j] = 0
				continue
			}
			ratio := float64(SoftWonStats[i][j]) / float64(denom)
			ratioSoftWon[i][j] = ratio
		}
	}

	for i := 2; i < len(ratioSoftDoubleWon); i++ {
		for j := 1; j < 11; j++ {
			denom := SoftDoubleWonStats[i][j] + SoftDoubleLostStats[i][j]
			if denom == 0 {
				ratioSoftDoubleWon[i][j] = 0
				continue
			}
			ratio := float64(SoftDoubleWonStats[i][j]) / float64(denom)
			ratioSoftDoubleWon[i][j] = ratio
		}
	}

	bufOutputFileWriter.WriteString("\n Ratio Won Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range ratioWon {
		if i < 2 {
			continue
		}
		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range ratioWon[i] {
			if j == 0 {
				continue
			}
			rowString := ""
			if stats < 1e-6 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6.3f", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a ratioWon row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Ratio Double Won Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range ratioDoubleWon {
		if i < 2 {
			continue
		}
		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range ratioDoubleWon[i] {
			if j == 0 {
				continue
			}
			rowString := ""
			if stats < 1e-6 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6.3f", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a ratioDoubleWon row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Ratio Soft Won Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range ratioSoftWon {
		if i < 2 {
			continue
		}
		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range ratioSoftWon[i] {
			if j == 0 {
				continue
			}
			rowString := ""
			if stats < 1e-6 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6.3f", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a ratioSoftWon row.  Error is", err)
		}
	}

	bufOutputFileWriter.WriteString("\n Ratio Soft Double Won Array \n")
	bufOutputFileWriter.WriteString("         A      2      3      4      5      6      7      8      9     10 \n")
	bufOutputFileWriter.WriteString("--------------------------------------------------------------------------\n")
	for i := range ratioSoftDoubleWon {
		if i < 2 {
			continue
		}
		s := fmt.Sprintf("%2d:", i)
		bufOutputFileWriter.WriteString(s)
		for j, stats := range ratioSoftDoubleWon[i] {
			if j == 0 {
				continue
			}
			rowString := ""
			if stats < 1e-6 {
				rowString = "     - "
			} else {
				rowString = fmt.Sprintf(" %6.3f", stats)
			}
			bufOutputFileWriter.WriteString(rowString)
		}
		_, err = bufOutputFileWriter.WriteRune('\n')
		if err != nil {
			fmt.Println(" Error while writing a ratioSoftDoubleWon row.  Error is", err)
		}
	}

	_, err = bufOutputFileWriter.WriteRune('\n')
	_, err = bufOutputFileWriter.WriteRune('\n')
	_, err = bufOutputFileWriter.WriteRune('\n')
	bufOutputFileWriter.Flush()
	OutputHandle.Close()
} // wrStatsToFile

// ------------------------------------------------------- main -----------------------------------
// ------------------------------------------------------- main -----------------------------------
func main() {
	fmt.Printf("BlackJack Simulation Prgram, written in Go.  Last altered %s, compiled by %s \n", lastAltered, runtime.Version())

	flag.BoolVar(&verboseFlag, "v", false, " Verbose mode")
	flag.Parse()

	const InputExtDefault = ".strat"
	const OutputExtDefault = ".results"

	if len(os.Args) < 2 {
		fmt.Printf(" Usage:  bj <strategy-file.%s> \n", InputExtDefault)
		os.Exit(1)
	}

	/* now handled by verbose mode.
	       fmt.Print(" Display each round? Y/n ")
	   	ans := ""
	   	_, e := fmt.Scanln(&ans)
	   	if e != nil {
	   		displayRound = false
	   	} else if ans == "y" {
	   		displayRound = true
	   	}
	*/
	displayRound = verboseFlag
	deck = make([]int, 0, NumOfCards)

	//commandline := getcommandline.GetCommandLineString()
	//commandline := os.Args[1] // Args[0] is the name of the executable binary
	commandLine := flag.Arg(0)
	BaseFilename := filepath.Clean(commandLine)
	var Filename string

	if strings.Contains(BaseFilename, ".") {
		Filename = BaseFilename
	} else {
		Filename = BaseFilename + InputExtDefault
	}
	if verboseFlag {
		fmt.Printf("\n Strategy matrix filename is %s. \n\n", Filename)
	}

	fi, err := os.Stat(Filename)
	if err != nil {
		fmt.Println(Filename, "does not exist.  Exiting.")
		os.Exit(1)
	}

	//byteslice := make([]byte, 0, FI.Size()+50)
	byteSlice, er := os.ReadFile(Filename)
	if er != nil {
		fmt.Println(" Error from os.ReadFile: ", er, ".  Exiting.")
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Printf(" Read %s but not yet processed; len(byteSlice) = %d, filesize = %d \n\n", fi.Name(), len(byteSlice), fi.Size())
	}

	bytesReader := bytes.NewReader(byteSlice) // NewReader does not allocate memory like NewBuffer does.

	if verboseFlag {
		pause()
	}

	ReadStrategyMatrix(bytesReader)
	if verboseFlag {
		fmt.Printf(" StrategyMatrix read and processed successfully.\n")
	}

	// Construct results filename to receive the results.
	OutputFilename = BaseFilename + OutputExtDefault
	OutputHandle, err = os.OpenFile(OutputFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Cound not write output file.  If on my Windows Desktop, its likely my security precautions are in effect and I have to let this pgm thru.  Exiting.")
		os.Exit(1)
	}
	defer OutputHandle.Close()
	bufOutputFileWriter = bufio.NewWriter(OutputHandle)
	defer bufOutputFileWriter.Flush()

	_, _ = bufOutputFileWriter.WriteString("==============================================================================\n")
	_, _ = bufOutputFileWriter.WriteString("==============================================================================\n")
	_, err = bufOutputFileWriter.WriteRune('\n')
	if err != nil {
		fmt.Println(" Writing to output file,", OutputFilename, "produced this error:", err, ".  Exiting")
		os.Exit(1)
	}

	date := time.Now()
	datestring := date.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.
	str := fmt.Sprintf(" Date is %s; Dealer hitting on soft 17 flag is %v, Re-split aces flag is %v \n \n",
		datestring, dealerHitsSoft17, resplitAcesFlag)

	_, err = bufOutputFileWriter.WriteString(str)

	WriteStrategyMatrix(bufOutputFileWriter)

	_, err = bufOutputFileWriter.WriteString("==============================================================================\n")
	_, err = bufOutputFileWriter.WriteRune('\n')
	if err != nil {
		fmt.Println(" Writing to output file,", OutputFilename, "produced this error:", err, ".  Exiting")
		os.Exit(1)
	}

	// just in case there is a panic of some type, this file will be closed so I can inspect it, so far.
	bufOutputFileWriter.Flush()
	OutputHandle.Close()

	// Init the deck and SurrenderStrategyMatrix, and shuffle the deck
	InitDeck()
	initSurrenderStrategyMatrix() // only used when the surrender option is wanted in the Strategy matrix.

	if verboseFlag {
		fmt.Println(" Initialized deck.  There are", len(deck), "cards in this deck.")
		fmt.Println(deck)
		fmt.Println()
	}

	t0 := time.Now()

	rand.Seed(int64(t0.Nanosecond()))

	//       need to shuffle here
	swapFnt := func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	}
	milliSec := date.Nanosecond() / 1e6
	for i := 0; i < milliSec; i++ { // increase the shuffling, since it's not so good, esp noticable when I'm using only 1 deck for testing of this.
		_ = rand.Int()
		rand.Shuffle(len(deck), swapFnt)
		//rand.Shuffle(len(deck), swapfnt)  I think this is too much.
	}
	timeToShuffle := time.Since(t0) // timeToShuffle is a Duration type, which is an int64 but has methods.
	if displayRound || verboseFlag {
		fmt.Println(" It took ", timeToShuffle.String(), " to shuffle this file.  millisec=", milliSec, ".")
		fmt.Println()
		fmt.Println(" Shuffled deck still has", len(deck), "cards.")
		fmt.Println(deck)
	}

	fmt.Print(" How many hands to play: ")
	_, err = fmt.Scanln(&numOfPlayers)
	if err != nil {
		numOfPlayers = 1
	}

	playerHand = make([]handType, 0, maxNumOfPlayers)
	runsWon = make([]int, 0, maxNumOfPlayers)
	runsLost = make([]int, 0, maxNumOfPlayers)
	/* Just a demo of what GoLand can do automatically
	   hand = handType{
	       card1:           0,
	       card2:           0,
	       total:           0,
	       doubledflag:     false,
	       surrenderedflag: false,
	       bustedflag:      false,
	       BJflag:          false,
	       softflag:        false,
	       result:          0,
	   }
	*/
	hand = handType{}
	for h := 0; h < numOfPlayers; h++ {
		playerHand = append(playerHand, hand)
	}

	if displayRound {
		fmt.Println(" Initial number of hands is", len(playerHand))
		fmt.Println()
	}

	/* {{{
			dealCards()
			fmt.Println(" after cards were first dealt.  Player(s) first")
			fmt.Println(playerHand)
			fmt.Println()
			fmt.Println(" Dealer last.")
			fmt.Println(dealerHand)
			fmt.Println()
	           }}} */
	t1 := time.Now()
	// Main loop of this simulator, to play all rounds
PlayAllRounds:
	for j := 0; j < maxNumOfHands; j++ {
		playAllHands()
		showDown()
		incrementStats()

		if displayRound {
			fmt.Printf("\n\n There are %d hand(s), including splits \n\n", len(playerHand))
			for i := range playerHand {
				fmt.Printf(" playerHand[%d]: card1=%d, card2=%d, total=%d, notAvirgin=%t, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, split=%t, BJ=%t, result=%s \n",
					i, playerHand[i].card1, playerHand[i].card2, playerHand[i].total, playerHand[i].notAvirginflag, playerHand[i].doubledflag, playerHand[i].surrenderedflag,
					playerHand[i].bustedflag, playerHand[i].pair, playerHand[i].softflag, playerHand[i].splitflag, playerHand[i].BJflag, resultNames[playerHand[i].result])
			}
			fmt.Printf(" dealerHand is card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, BJ=%t \n",
				dealerHand.card1, dealerHand.card2, dealerHand.total, dealerHand.doubledflag, dealerHand.surrenderedflag,
				dealerHand.bustedflag, dealerHand.pair, dealerHand.softflag, dealerHand.BJflag)
			fmt.Println(" ------------------------------------------------------------------")
			fmt.Println()
			fmt.Print(" Continue? Y/n:  Stop or Exit also work.  ")
			fmt.Println()
			var ans string
			_, err := fmt.Scanln(&ans)
			if err != nil {
				ans = ""
			}
			if ans == "n" || ans == "stop" || ans == "exit" || ans == "q" {
				break PlayAllRounds
			}
			fmt.Println(" ------------------------------------------------------------------")
			clearscreen[runtime.GOOS]()
		}

		// Need to remove splits, if any, from the player hand slice.
		playerHand = playerHand[:numOfPlayers]

		if currentCard > len(deck)*3/4 { // shuffle if 3/4 of the deck has been played thru.
			doTheShuffle()
			if displayRound {
				fmt.Println(" shuffling ...")
			}
		}
		if displayRound {
			fmt.Println(" deck current position is, possibly after a shuffle.", currentCard)
			fmt.Printf(" %d hands were planned; %d were actually played. \n\n", j, totalHands)
		}
		//		if j > 100_000 {  Not needed.  It seems it takes ~ 1/3 - 1/2 sec to run thru 1 million hands.
		//			break         A lot has changed over 20 years, when it took about 20 min to run thru 1 million hands, IIRC.
		//		}                 On z76, 1 billion took ~5 min.  But then the stats were each 6 figures, so I would need to make this change if I do that.
		//                        On z76, 10 million hands took ~2 sec.
	}

	elapsed := time.Since(t1)
	elapsedString := fmt.Sprintf(" Playing %d hands took %s \n", totalHands, elapsed.String())
	fmt.Println(elapsedString)

	// time for the stats.
	// need to remember to display totalsplits, totaldoubles, totalbusts, totalHands.
	// need to calculate a run, how long it ran, and display these.  That's in the runs slice of int.

	OutputHandle, err = os.OpenFile(OutputFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Could not write output file.  If on my Windows Desktop, likely my security precautions in effect and I have to let this pgm thru.  Exiting.")
		os.Exit(1)
	}
	bufOutputFileWriter = bufio.NewWriter(OutputHandle)
	defer bufOutputFileWriter.Flush()
	defer OutputHandle.Close()

	score = 1.5*float64(totalBJwon) + float64(totalDblWins)*2 + float64(totalWins) - float64(totalDblLosses)*2 - float64(totalLosses) -
		float64(totalSurrenders)/2
	scoreString := fmt.Sprintf(" Score=  %.2f, BJ won=%d, wins=%d, losses=%d, Double wins=%d, Double losses=%d, surrendered=%d \n",
		score, totalBJwon, totalWins, totalLosses, totalDblWins, totalDblLosses, totalSurrenders)
	fmt.Println(scoreString)

	var ratioTotalDblWins, ratioTotalWins, ratioTotalDblLosses, ratioTotalLosses float64
	ratioTotalDblWins = float64(totalDblWins) / float64(totalDblWins+totalDblLosses)
	ratioTotalDblLosses = float64(totalDblLosses) / float64(totalDblWins+totalDblLosses)
	ratioTotalWins = float64(totalWins) / float64(totalWins+totalLosses)
	ratioTotalLosses = float64(totalLosses) / float64(totalWins+totalLosses)
	ratioScore := 100 * score / float64(totalHands)
	ratioString := fmt.Sprintf(" RatioScore= %.4f%%,  TotalWins= %.4f, TotalLosses= %.4f, TotalDblWins= %.4f, TotalDblLosses= %.4f \n",
		ratioScore, ratioTotalWins, ratioTotalLosses, ratioTotalDblWins, ratioTotalDblLosses)
	fmt.Println(ratioString)
	bufOutputFileWriter.WriteString(ratioString)

	bufOutputFileWriter.WriteString(elapsedString)
	bufOutputFileWriter.WriteString(scoreString)

	totalHandsFloat := float64(totalHands)
	totalBJhandFloat := float64(totalBJwon + totalBJpushed)
	ratioBJwon := float64(totalBJwon) / totalBJhandFloat
	ratioBJpushed := float64(totalBJpushed) / totalBJhandFloat
	ratioHandsWon := float64(totalWins) / totalHandsFloat
	ratioBJdealerAce := float64(totalBJwithDealerAce) / totalBJhandFloat
	ratioBusts := float64(totalBusts) / totalHandsFloat
	ratioSplits := float64(totalSplits) / totalHandsFloat
	outputratiostring := fmt.Sprintf(" ratio BJ won= %.3f, ratio BJ pushed= %.3f, BJ w/ dealer Ace = %d,  ratio BJ with dlr Ace= %.4f \n\n", ratioBJwon, ratioBJpushed, totalBJwithDealerAce, ratioBJdealerAce)
	fmt.Print(outputratiostring)
	bufOutputFileWriter.WriteString(outputratiostring)
	outputratiostring = fmt.Sprintf(
		" ratio Hands Won/total hands= %.3f, total busts= %d, ratio Busts/total hands= %.3f, total splits= %d, ratio splits= %.4f \n\n",
		ratioHandsWon, totalBusts, ratioBusts, totalSplits, ratioSplits)
	fmt.Print(outputratiostring)
	bufOutputFileWriter.WriteString(outputratiostring)

	sort.Sort(sort.Reverse(sort.IntSlice(runsWon)))
	sort.Sort(sort.Reverse(sort.IntSlice(runsLost)))
	runswonstring := fmt.Sprintf(" runs of won hands: %v \n", runsWon[:25]) // Else there are too many of them
	runsloststring := fmt.Sprintf(" runs of lost hands: %v \n", runsLost[:25])
	bufOutputFileWriter.WriteString(runswonstring)
	bufOutputFileWriter.WriteString(runsloststring)

	runswonstring = fmt.Sprintf(" runs of won hands: %v \n", runsWon[:20])
	runsloststring = fmt.Sprintf(" runs of lost hands: %v \n", runsLost[:20])
	fmt.Println(runswonstring)
	fmt.Println(runsloststring)

	wrStatsToFile()

} // main

// ----------------------------------------------------------
// readLine

func readLine(r *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		byt, err := r.ReadByte() // byte is a reserved word for a variable type.
		/*		if verboseFlag {
					fmt.Printf(" %c %v ", byt, err)
					pause()
				}
		*/ //if err == io.EOF {  I have to return io.EOF so the EOF will be properly detected as such.
		//	return strings.TrimSpace(sb.String()), nil
		//} else
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
		if byt == '\n' { // will stop scanning a line after seeing these characters like in bash or C-ish.
			return strings.TrimSpace(sb.String()), nil
		}
		if byt == '\r' {
			continue
		}
		if byt == '#' || byt == '/' {
			discardRestOfLine(r)
			return strings.TrimSpace(sb.String()), nil
		}
		err = sb.WriteByte(byt)
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
	}
} // readLine
// ----------------------------------------------------------------------
func discardRestOfLine(r *bytes.Reader) { // To allow comments on a line, I have to discard rest of line from the bytes.Reader
	for { // keep swallowing characters until EOL or an error.
		rn, _, err := r.ReadRune()
		if err != nil {
			return
		}
		if rn == '\n' {
			return
		}
	}
}

// ----------------------------------------------------------------------

func pause() {
	fmt.Printf(" hit any key to continue   ")
	var ans string
	fmt.Scanln(&ans)
	fmt.Printf("%s\n", ans)
}
