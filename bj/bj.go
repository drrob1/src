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
*/
import (
	"bufio"
	"bytes"
	"fmt"
	"getcommandline"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"tknptr"
)

const lastAltered = "June 10, 2020"

/*
  REVISION HISTORY
  ======== =======
   4 Jun 20 -- Started to convert the old blackjack Modula-2 code to Go.  This will take a while.

*/

var OptionName = []string{"S  ", "H  ", "D  ", "SP ", "SUR"} // Stand, Hit, Double, Split, Surrender

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

var Strategy [22]OptionRowType // Modula-2 ARRAY [5..21] OF OptionRowType.  I'm going to ignore rows that are not used.

// Because of change in logic, SoftStrategy needs enough rows to handle more cases.  Maybe.  I changed the soft hand logic also.
var SoftStrategy [22]OptionRowType // Modula-2 ARRAY [2..11] of OptionRowType.  Also going to ignore rows that are not used.

var PairStrategy [11]OptionRowType // Modula-2 ARRAY [1..10] of OptionRowType.  Same about unused rows.
//var StrategyErrorFlag bool         // not sure if I'll need this yet.

const numOfDecks = 8

//const numOfDecks = 1 // for testing of shuffles.  When I'm confident shuffling works correctly, I'll return it to 8 decks.
const maxNumOfPlayers = 100
const maxNumOfHands = 1_000_000_000 // 1 million, for now.
const NumOfCards = 52 * numOfDecks

var resultNames = []string{"  lost", "  pushed", "  won", "  surrend", "  LostDbl", "  WonDbl", "  LostToBJ", "  PushedBJ", "  WonBJ"}

type handType struct {
	card1, card2, total                                                         int
	doubledflag, surrenderedflag, bustedflag, pair, softflag, splitflag, BJflag bool
	result                                                                      int
}

var displayRound, resplitAcesFlag, lastHandWinLoseFlag, dealerHitsSoft17 bool

var playerHand []handType
var hand handType
var dealerHand handType
var numOfPlayers, currentCard int
var deck []int
var prevResult []int
var runs []int
var totalWins, totalLosses, totalPushes, totalDblWins, totalDblLosses, totalBJwon, totalBJpushed, totalBJwithDealerAce, totalSplits,
	totalDoubles, totalSurrenders, totalBusts, totalHands int
var winsInARow, lossesInARow int
var score float64
var clearscreen map[string]func()

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

// ------------------------------------------------------- ReadStrategy -----------------------------------
func ReadStrategy(buf *bytes.Buffer) {
	for {
		rowbuf, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(" Error from bytes.Buffer ReadString is", err)
			os.Exit(1)
		}
		rowbuf = strings.TrimSpace(rowbuf)
		if len(rowbuf) == 0 { // ignore blank lines
			continue
		}
		tknbuf := tknptr.NewToken(rowbuf)
		rowID, EOL := tknbuf.GetToken(true) // force upper case token.
		if EOL {
			return
		}
		row := make([]int, 0, 10) // a single strategy row.
		for {
			token, eol := tknbuf.GetToken(true) // force upper case token
			if eol {
				break
			}
			if EOL || token.State != tknptr.ALLELSE {
				return
			}
			i := GetOption(token)
			if i == ErrorValue { // allow for comments after the strategy codes on same line.
				return
			}
			row = append(row, i)
		}

		if rowID.Isum >= 4 && rowID.Isum <= 21 { // assign Strategy codes to proper row in Strategy Decision Matrix
			Strategy[rowID.Isum] = row
		} else if rowID.State == tknptr.DGT {
			switch rowID.Isum {
			case 22:
				PairStrategy[2] = row
			case 33:
				PairStrategy[3] = row
			case 44:
				PairStrategy[4] = row
			case 55:
				PairStrategy[5] = row
			case 66:
				PairStrategy[6] = row
			case 77:
				PairStrategy[7] = row
			case 88:
				PairStrategy[8] = row
			case 99:
				PairStrategy[9] = row
			case 1010:
				PairStrategy[10] = row
			default:
				fmt.Println(" Invalid Pair Row value:", rowID) // rowID is a struct, so all of it will be output.
				fmt.Print(" continue? y/n ")
				ans := ""
				_, _ = fmt.Scanln(&ans)
				ans = strings.ToLower(ans)
				if ans != "y" {
					os.Exit(1)
				}
			} // end switch-case
		} else if rowID.State == tknptr.ALLELSE {
			if rowID.Str == "AA" {
				PairStrategy[1] = row
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
				SoftStrategy[i] = row
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
				SoftStrategy[i+1] = row
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
} // ReadStrategy

// ------------------------------------------------------- WriteStrategy -----------------------------------
func WriteStrategy(filehandle *bufio.Writer) {
	filehandle.WriteString(" Regular Strategy Matrix: \n")

	// First write out regular Strategy
	for i, row := range Strategy {
		if i < 5 { // ignore rows < 5, as these are special cases and are in the other matrixes
			continue
		}
		outputline := fmt.Sprintf(" %d: ", i)
		for _, j := range row {
			s := fmt.Sprintf("%s  ", OptionName[j])
			outputline += s
		}
		filehandle.WriteString(outputline)
		filehandle.WriteRune('\n')
	}

	// Now write out Soft Strategy
	filehandle.WriteString(" \n \n Soft Strategy Matrix, IE, Strategy for soft hands: \n")
	for i, row := range SoftStrategy {
		if i < 3 { // ignore row 0, 1 and 2, as these are not a valid blackjack hand or are in PairStrategy
			continue
		}
		outputline := fmt.Sprintf(" s%d: ", i)
		for _, j := range row {
			s := fmt.Sprintf("%s  ", OptionName[j])
			outputline += s
		}
		filehandle.WriteString(outputline)
		filehandle.WriteRune('\n')
	}

	// Now write out Pair Strategy
	filehandle.WriteString(" \n \n Pair Strategy Matrix: \n")
	for i, row := range PairStrategy {
		if i < 1 { // ignore row 0 as it is not a valid blackjack hand
			continue
		}
		outputline := fmt.Sprintf(" %d-%d: ", i, i)
		for _, j := range row {
			s := fmt.Sprintf("%s  ", OptionName[j])
			outputline += s
		}
		filehandle.WriteString(outputline)
		filehandle.WriteRune('\n')
	}

	filehandle.WriteRune('\n')
	filehandle.WriteRune('\n')
	filehandle.WriteRune('\n')
} // WriteStrategy

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
	fmt.Printf(" Entering DealerHit.  Hand is: %d %d, total=%d \n", dealerHand.card1, dealerHand.card2, dealerHand.total)
	if dealerHand.BJflag {
		return
	}
	for dealerHand.total < 18 { // always exit if >= 18.  Loop if < 18.
		fmt.Printf("\n Top of for Dealerhand total.  Hand is: %d %d, total=%d, soft=%t \n",
			dealerHand.card1, dealerHand.card2, dealerHand.total, dealerHand.softflag)
		if dealerHand.softflag && dealerHand.total <= 11 && dealerHitsSoft17 {
			if dealerHand.total > 7 && dealerHand.total <= 11 {
				dealerHand.total += 10
			}
			if dealerHand.total <= 17 {
				fmt.Printf(" Dealer soft17 hit loop.  Total=%d \n", dealerHand.total)
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
				fmt.Printf(" Dealer soft no hit soft17 loop.  Total=%d \n", dealerHand.total)
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
				fmt.Printf(" Dealer hard hand hit loop.  Total=%d \n", dealerHand.total)
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
		fmt.Printf(" Top of hitMePlayer: playerHand[%d]: card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, BJ=%t \n\n",
			i, playerHand[i].card1, playerHand[i].card2, playerHand[i].total, playerHand[i].doubledflag, playerHand[i].surrenderedflag,
			playerHand[i].bustedflag, playerHand[i].pair, playerHand[i].softflag, playerHand[i].BJflag)

		if playerHand[i].softflag && playerHand[i].total <= 11 { // if total > 11, Ace can only be 1 and not a 10.
			strategy := SoftStrategy[playerHand[i].total][dealerHand.card1-1]
			if strategy == Double && cannotDouble { // can only double on first time thru, else have more than 2 cards.
				strategy = Hit
			} else {
				cannotDouble = true // so next time thru the loop cannot double
			}
			fmt.Printf(" SoftStrategy[%d][%d] = %s \n\n", playerHand[i].total, dealerHand.card1-1, OptionName[strategy])
			switch strategy {
			case Stand:
				if playerHand[i].total <= 11 { // here is the logic of a soft hand
					playerHand[i].total += 10
				}
				break MainLoop

			case Hit:
				card := getCard()
				playerHand[i].total += card
				if playerHand[i].total > 21 { // if you hit and bust a soft hand, then the Ace must be 1
					playerHand[i].bustedflag = true
					break MainLoop
				}
				cannotDouble = true

			case Double: // here this means hit only once.
				if !playerHand[i].doubledflag {
					playerHand[i].total += getCard()
					if playerHand[i].total > 21 { // since Ace is initially treated as a 1, this should never bust.  Unless doubling 12+
						playerHand[i].bustedflag = true
					} else if playerHand[i].total <= 11 { // soft hand effect.
						playerHand[i].total += 10
					}
				}
				playerHand[i].doubledflag = true
				break MainLoop

			case Surrender:
				fmt.Printf(" in hitMePlayer for soft hands and got a surrender option.  I=%d, hand=%v \n", i, playerHand[i])
				return

			case ErrorValue:
				fmt.Printf(" in hitMePlayer and got errorValue.  I is %d, and hand is %v \n", i, playerHand[i])
				return
			} // switch-case
		} else { // not a soft hand where Ace can represent 11.
			strategy := Strategy[playerHand[i].total][dealerHand.card1-1]
			if strategy == Double && cannotDouble { // can only double on first time thru, else have more than 2 cards.
				strategy = Hit
			} else {
				cannotDouble = true // so next time thru the loop cannot double
			}
			fmt.Printf(" Strategy[%d][%d] = %s \n\n", playerHand[i].total, dealerHand.card1-1, OptionName[strategy])
			switch strategy {
			case Stand:
				fmt.Printf(" Hard HitMe - stand.  total= %d \n", playerHand[i].total)
				return

			case Hit:
				fmt.Printf(" Hard HitMe - hit.  total= %d \n", playerHand[i].total)
				newcard := getCard()
				playerHand[i].total += newcard
				if newcard == 1 { // if pulled an Ace, next time around this is considered a soft hand.
					playerHand[i].softflag = true
				}
				if playerHand[i].total > 21 {
					playerHand[i].bustedflag = true
					return
				} else if playerHand[i].softflag && playerHand[i].total >= 9 && playerHand[i].total <= 11 {
					playerHand[i].total += 10
				}

			case Double: // hit only once.
				fmt.Printf(" Hard HitMe - double.  total= %d \n", playerHand[i].total)
				if !playerHand[i].doubledflag { // must only draw 1 card.
					newcard := getCard()
					playerHand[i].total += newcard
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

			case Surrender:
				fmt.Printf(" Hard HitMe - surrender.  total= %d \n", playerHand[i].total)
				playerHand[i].surrenderedflag = true
				return

			case ErrorValue:
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
	fmt.Printf("\n Entering splitHand.  card1=%d, card2=%d \n", playerHand[i].card1, playerHand[i].card2)
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

	fmt.Printf(" splitHand: 1st hand.card1=%d, 1st hand.card2=%d, 1st hand.total=%d; split-off hnd.card1=%d, hnd.card2=%d, hnd.total=%d \n\n ",
		playerHand[i].card1, playerHand[i].card2, playerHand[i].total, hnd.card1, hnd.card2, hnd.total)

	playerHand = append(playerHand, hnd) // can't get blackjack after a split.
	fmt.Println(" Exiting splitHand: playerHand slice =", playerHand)
} // splitHand

// ------------------------------------------------------- playAhand -----------------------------------
func playAhand(i int) {
	fmt.Printf(" Top of playAhand: playerHand[%d]: card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, BJ=%t \n",
		i, playerHand[i].card1, playerHand[i].card2, playerHand[i].total, playerHand[i].doubledflag, playerHand[i].surrenderedflag,
		playerHand[i].bustedflag, playerHand[i].pair, playerHand[i].softflag, playerHand[i].BJflag)
	fmt.Printf(" Top of playAhand: dealerHand is card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, BJ=%t \n",
		dealerHand.card1, dealerHand.card2, dealerHand.total, dealerHand.doubledflag, dealerHand.surrenderedflag,
		dealerHand.bustedflag, dealerHand.pair, dealerHand.softflag, dealerHand.BJflag)
	fmt.Println()
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
			fmt.Println(" About to consider splitting aces.  Have not yet checked to see about resplitting aces.")
			if playerHand[i].splitflag && !resplitAcesFlag { // not allowed to resplit aces (or double the hand).
				return
			}
			splitHand(i)
			playAhand(i) // I'm trying out recursion to solve this problem.  The split-off additional hand will be handled by playAllHands.
			return
		} else {
			strategy := PairStrategy[playerHand[i].card1][dealerHand.card1-1]
			fmt.Printf("playAhand for hand=%d, PairStrategy[%d][%d] = %s \n\n", i, playerHand[i].total, dealerHand.card1-1, OptionName[strategy])
			switch strategy {
			case Stand:
				return
			case Hit:
				hitMePlayer(i)
			case Double:
				// playerHand[i].doubledflag = true  This prevents taking a card in HitMePlayer.  I'll let Strategy for 10 take over here.
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
				fmt.Printf(" PairStrategy returned ErrorValue.  i=%d, hand= %v \n", i, playerHand[i])
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
		default:
			if playerHand[i].bustedflag {
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
				} else {
					playerHand[i].result = won
					totalWins++
				}
			} else if playerHand[i].total > dealerHand.total {
				if playerHand[i].doubledflag {
					playerHand[i].result = wondbl
					totalDblWins++
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
				} else {
					playerHand[i].result = lost
					totalLosses++
				}
			}
		} // seitch-case
	} // for range over all hands, incl'g split hands.
} // showDown

// ------------------------------------------------------- main -----------------------------------
// ------------------------------------------------------- main -----------------------------------
func main() {
	fmt.Printf("BlackJack Simulation Prgram, written in Go.  Last altered %s \n", lastAltered)

	InputExtDefault := ".strat"
	OutputExtDefault := ".results"

	if len(os.Args) < 2 {
		fmt.Printf(" Usage:  bj <strategy-file.%s> \n", InputExtDefault)
		os.Exit(1)
	}

	fmt.Print(" Display each round? Y/n ")
	ans := ""
	_, e := fmt.Scanln(&ans)
	if e != nil {
		displayRound = true
	} else if ans == "y" {
		displayRound = true
	}

	/*  These are now set in the .strat file.
	    fmt.Print(" Simulate dealer hitting on a soft 17? y/n ")
	    ans := ""
	    _, _ = fmt.Scanln(&ans)
	    ans = strings.ToLower(ans)
	    if ans == "y" {
	    	DealerHitsSoft17 = true
	    }
	    fmt.Println(" Value of Dealer Hit Soft 17 flag is", DealerHitsSoft17)

	    fmt.Print(" Allow the re-splitting of aces? y/n ")
	    _, _ = fmt.Scanln(&ans)
	    ans = strings.ToLower(ans)
	    if ans == "y" {
	    	ResplitAces = true
	    }
	    fmt.Println(" Value of Re-split aces flag is", ResplitAces)
	*/

	deck = make([]int, 0, NumOfCards)

	commandline := getcommandline.GetCommandLineString()
	BaseFilename := filepath.Clean(commandline)
	Filename := ""

	if strings.Contains(BaseFilename, ".") {
		Filename = BaseFilename
	} else {
		Filename = BaseFilename + InputExtDefault
	}

	FI, err := os.Stat(Filename)
	if err != nil {
		fmt.Println(Filename, "does not exist.  Exiting.")
		os.Exit(1)
	}

	byteslice := make([]byte, 0, FI.Size()+50)
	byteslice, err = ioutil.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error from ioutil.ReadFile: ", err, ".  Exiting.")
		os.Exit(1)
	}

	bytesbuffer := bytes.NewBuffer(byteslice)

	ReadStrategy(bytesbuffer)

	// Construct results filename to receive the results.
	OutputFilename := BaseFilename + OutputExtDefault
	OutputHandle, err := os.OpenFile(OutputFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Cound not write output file.  If on my Windows Desktop, likely my security precautions in effect and I have to let this pgm thru.  Exiting.")
		os.Exit(1)
	}
	bufOutputFileWriter := bufio.NewWriter(OutputHandle)
	defer bufOutputFileWriter.Flush()
	defer OutputHandle.Close()

	date := time.Now()
	datestring := date.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.
	str := fmt.Sprintf(" Date is %s; Dealer hitting on soft 17 flag is %v, Re-split aces flag is %v \n \n",
		datestring, dealerHitsSoft17, resplitAcesFlag)

	_, err = bufOutputFileWriter.WriteString(str)

	WriteStrategy(bufOutputFileWriter)

	_, err = bufOutputFileWriter.WriteString("------------------------------------------------------\n")
	_, err = bufOutputFileWriter.WriteRune('\n')
	if err != nil {
		fmt.Println(" Writing to output file,", OutputFilename, "produced this error:", err, ".  Exiting")
		os.Exit(1)
	}

	// just in case there is a panic of some type, this file will be closed so I can inspect it,
	// so far.
	bufOutputFileWriter.Flush()
	OutputHandle.Close()

	// Init and shuffle the deck
	InitDeck()
	/* */
	fmt.Println(" Initialized deck.  There are", len(deck), "cards in this deck.")
	fmt.Println(deck)
	fmt.Println()
	/* */

	t0 := time.Now()

	rand.Seed(int64(time.Now().Nanosecond()))
	//       need to shuffle here
	swapfnt := func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	}
	millisec := date.Nanosecond() / 1e6
	for i := 0; i < millisec; i++ { // increase the shuffling, since it's not so good, esp noticable when I'm using only 1 deck for testing of this.
		rand.Shuffle(len(deck), swapfnt)
		rand.Shuffle(len(deck), swapfnt)
	}
	timeToShuffle := time.Since(t0) // timeToShuffle is a Duration type, which is an int64 but has methods.
	fmt.Println(" It took ", timeToShuffle.String(), " to shuffle this file.  Or", timeToShuffle.Nanoseconds(), "ns to shuffle.")
	fmt.Println()
	/* */
	fmt.Println(" Shuffled deck still has", len(deck), "cards.")
	fmt.Println(deck)
	/* */

	fmt.Print(" How many hands to play: ")
	_, err = fmt.Scanln(&numOfPlayers)
	if err != nil {
		numOfPlayers = 1
	}

	playerHand = make([]handType, 0, maxNumOfPlayers)
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

	fmt.Println(" Initial number of hands is", len(playerHand))
	fmt.Println()
	/*
		dealCards()
		fmt.Println(" after cards were first dealt.  Player(s) first")
		fmt.Println(playerHand)
		fmt.Println()
		fmt.Println(" Dealer last.")
		fmt.Println(dealerHand)
		fmt.Println()
	*/

	// Main loop of this simulator, to play all rounds
PlayAllRounds:
	for j := 0; j < maxNumOfHands; j++ {
		playAllHands()
		showDown()

		if displayRound {
			fmt.Printf("\n\n There are %d hand(s), including splits \n\n", len(playerHand))
			for i := range playerHand {
				fmt.Printf(" playerHand[%d]: card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, BJ=%t, result=%s \n",
					i, playerHand[i].card1, playerHand[i].card2, playerHand[i].total, playerHand[i].doubledflag, playerHand[i].surrenderedflag,
					playerHand[i].bustedflag, playerHand[i].pair, playerHand[i].softflag, playerHand[i].BJflag, resultNames[playerHand[i].result])
			}
			fmt.Printf(" dealerHand is card1=%d, card2=%d, total=%d, dbldn=%t, sur=%t, busted=%t, pair=%t, soft=%t, BJ=%t \n",
				dealerHand.card1, dealerHand.card2, dealerHand.total, dealerHand.doubledflag, dealerHand.surrenderedflag,
				dealerHand.bustedflag, dealerHand.pair, dealerHand.softflag, dealerHand.BJflag)
			fmt.Println(" ------------------------------------------------------------------")
			fmt.Println()
			fmt.Print(" Continue? Y/n:  Stop or Exit also work.  ")
			fmt.Println()
			_, err := fmt.Scanln(&ans)
			if err != nil {
				ans = ""
			}
			if ans == "n" || ans == "stop" || ans == "exit" {
				break PlayAllRounds
			}
			fmt.Println(" ------------------------------------------------------------------")
			clearscreen[runtime.GOOS]()
		}

		// Need to remove splits, if any, from the player hand slice.
		playerHand = playerHand[:numOfPlayers]

		fmt.Println(" deck current position is", currentCard)
		if currentCard > len(deck)*3/4 { // shuffle if 3/4 of the deck has been played thru.
			doTheShuffle()
			if displayRound {
				fmt.Println(" shuffling ...")
			}
		}
	}

	// time for the stats.  I have a fair amount of work to do for this.  Maybe I'll first test the basic BJ logic before I worry about
	// stats.

	score = 1.5*float64(totalBJwon) + float64(totalDblWins)*2 + float64(totalWins) - float64(totalDblLosses)*2 - float64(totalLosses) -
		float64(totalSurrenders)/2
	fmt.Printf(" Score=  %.2f, BJ won=%d, Double wins=%d, wins=%d, double losses=%d, losses=%d, surrendered=%d \n\n",
		score, totalBJwon, totalDblWins, totalWins, totalDblLosses, totalLosses, totalSurrenders)

} // main
