package main

/*
  BlackJack Simulator.
  Translated from Modula-2 that I wrote ca 95, and then converted to Windows in 2005.

  The strategy matrix is read in and consists of columns 1 .. 10 (here, 0 .. 9) where ace is first, and last is all 10 value cards.
  Each row begins w/ a hand total, and means what strategy to follow when your hand totals that hand total.
  These are the integers from 5 .. 21, then S2 .. S10, meaning soft 2 thru soft 10.  Soft 2 would be 2 aces, but since that's a pair,
  this row is ignored.  Then we have the pairs AA thru 99 and 1010.
  Since I'm dealing w/ integers, I'll ignore row indicies that are not convenient.  IE, ignore Strategy rows < 5, etc.

  The input file now will have a .strat extension, just to be clear.  And the output file will have same basefilename with .results extension.

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
	"path/filepath"
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

const Ace = iota
const (
	Stand = iota
	Hit
	Double
	Split
	Surrender
	ErrorValue
)

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

type OptionRowType []int // first element is for the ace, last element is for all 10 value cards (10, jack, queen, king)
// by making this a slice, I can append the rows as I read them from the input strategy file.

var Strategy [22]OptionRowType     // Modula-2 ARRAY [5..21] OF OptionRowType.  I'm going to ignore rows that are not used.
var SoftStrategy [12]OptionRowType // Modula-2 ARRAY [2..11] of OptionRowType.  Also going to ignore rows that are not used.
var PairStrategy [11]OptionRowType // Modula-2 ARRAY [1..10] of OptionRowType.  Same about unused rows.
var StrategyErrorFlag bool         // not sure if I'll need this yet.

const numOfDecks = 8
const maxNumOfPlayers = 7
const maxNumOfHands = 1_000_000_000 // 1 million, for now.
const HandsPerPlayer = 7            // I guess this means splitting hands, which can get crazy.
const NumOfCards = 52 * numOfDecks

var resultNames = []string{"  lost", "  pushed", "  won", "  surrend", "  LostDbl", "  WonDbl", "  LostToBJ", "  PushedBJ", "  WonBJ"}

type handType struct {
	card1, card2, total                                        int
	doubledflag, surrenderedflag, bustedflag, BJflag, softflag bool
	result                                                     int
}

var resplitAcesFlag, lastHandWinLoseFlag, readyToShuffleFlag bool

var player []handType
var dealer handType
var splitsArray []int // well, slice, actually.  But nevermind this.
var prevResult []int
var numOfPlayers int
var totalWins, totalLosses, totalPushes, totalDblWins, totalDblLosses, totalBJwon, totalBJpushed, totalBJwithDealerAce, totalSplits,
	totalDoubles, totalSurrenders, totalBusts, totalHands int
var score, winsInARow, lossesInARow int
var runs []int
var DealerHitsSoft17, ResplitAces bool
var deck []int

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

func ReadStrategy(buf *bytes.Buffer) {
	for {
		rowbuf, err := buf.ReadString('\n')
		if err == io.EOF{
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

		if rowID.Isum >= 5 && rowID.Isum <= 21 { // assign Strategy codes to proper row in Strategy Decision Matrix
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
				fmt.Println(" Invalid Row value:", rowID) // rowID is a struct, so all of it will be output.
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
					fmt.Println(" Error in reading soft hand designation.")
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
				s := rowID.Str[1:]        // chop off the "A" and convert rest to int
				i, err := strconv.Atoi(s) // i is off by one b/o value of Ace is 11, not 10.
				if err != nil {
					fmt.Println(" Error in reading soft hand designation.")
					fmt.Print(" continue? y/n ")
					ans := ""
					_, _ = fmt.Scanln(&ans)
					ans = strings.ToLower(ans)
					if ans != "y" {
						os.Exit(1)
					}
				}
				SoftStrategy[i+1] = row
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
}

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
}

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

func main() {
	fmt.Printf("BlackJack Simulation Prgram, written in Go.  Last altered %s \n", lastAltered)

    InputExtDefault := ".strat"
    OutputExtDefault := ".results"

    if len(os.Args) < 2 {
		fmt.Printf(" Usage:  bj <strategy-file.%s> \n", InputExtDefault)
		os.Exit(1)
	}

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
    str := fmt.Sprintf(" Date is %s; Dealer hitting on soft 17 flag is %v, Re-split aces flag is %v \n \n", datestring, DealerHitsSoft17, ResplitAces)

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

    fmt.Println(" Initialized deck.  There are", len(deck), "cards in this deck.")
	fmt.Println(deck)
	fmt.Println()

    t0 := time.Now()

//       need to shuffle here
    swapfnt := func(i, j int) {
        deck[i], deck[j] = deck[j], deck[i]
    }
    rand.Shuffle(len(deck), swapfnt)
    rand.Shuffle(len(deck), swapfnt)

    timeToShuffle := time.Since(t0) // timeToShuffle is a Duration type, which is an int64 but has methods.
    fmt.Println(" It took ", timeToShuffle.String(), " to shuffle this file.  Or", timeToShuffle.Nanoseconds(),"ns to shuffle.")
    fmt.Println()

    fmt.Println(" Shuffled deck still has", len(deck), "cards.")
    fmt.Println(deck)

}
