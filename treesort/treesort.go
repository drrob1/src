package main

/*
  REVISION HISTORY
  ----------------
  17 Jul 17 -- Copied the modula-2 code here and will start to do this in Go.  I may not have looked at this since the 90's.
  20 May 20 -- Only converting to Go treesort, as the others I've already converted in the file now called mysorts.go.
                 Doesn't work.  I'm not debugging this.
*/

import (
	"bufio"
	"bytes"
	"fmt"
	"getcommandline"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const LastAltered = "May 20, 2020"

//   math.Mod and math.Remainder both return a float64.  math.Modf returns 2 float64 by separating the integer part and the fractional part.
//   math.Floor, math.Ceil

// EXPORT QUALIFIED MAXDIM,ITEMARRAY,LINKS,TREESORT,QUICKSORT,BININSSORT,INDIRQUICKSORT,INDIRBININSSORT,mergesort,heapsort;

//-----------------------------------------------------------------------+
//                               MAIN PROGRAM                            |
//-----------------------------------------------------------------------+

func main() {

	var filesize int64
	fmt.Println(" Sort a slice of strings, using only treesort.  Last altered", LastAltered)
	fmt.Println()

	// File I/O.  Construct filenames
	if len(os.Args) <= 1 {
		fmt.Println(" Usage: treesort <filename>")
		os.Exit(0)
	}

	Ext1Default := ".dat"
	Ext2Default := ".txt"
	OutDefault := ".sorted"

	date := time.Now()
	datestring := date.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.

	commandline := getcommandline.GetCommandLineString()
	BaseFilename := filepath.Clean(commandline)
	Filename := ""
	FileExists := false

	if strings.Contains(BaseFilename, ".") {
		Filename = BaseFilename
		FI, err := os.Stat(Filename)
		if err == nil {
			FileExists = true
			filesize = FI.Size()
		}
	} else {
		Filename = BaseFilename + Ext1Default
		FI, err := os.Stat(Filename)
		if err == nil {
			FileExists = true
			filesize = FI.Size()
		} else {
			Filename = BaseFilename + Ext2Default
			FI, err := os.Stat(Filename)
			if err == nil {
				FileExists = true
				filesize = FI.Size()
			}
		}
	}

	if !FileExists {
		fmt.Println(" File ", BaseFilename, " or ", Filename, " does not exist.  Exiting.")
		os.Exit(1)
	}

	byteslice := make([]byte, 0, filesize+5) // add 5 just in case
	byteslice, err := ioutil.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error from ioutil.ReadFile when reading ", Filename, ".  Exiting.")
		os.Exit(1)
	}

	bytesbuffer := bytes.NewBuffer(byteslice)

	OutFilename := BaseFilename + OutDefault
	//	OutputFile, err := os.Create(OutFilename)
	OutputFile, err := os.OpenFile(OutFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer OutputFile.Close()
	OutBufioWriter := bufio.NewWriter(OutputFile)
	defer OutBufioWriter.Flush()
	_, err = OutBufioWriter.WriteString("------------------------------------------------------\n")
	_, err = OutBufioWriter.WriteString(datestring)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	// Read in the words to sort
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Enter number of words for this run.  0 means full file: ")
	scanner.Scan()
	answer := scanner.Text()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
	requestedwordcount, err := strconv.Atoi(answer)
	if err != nil {
		fmt.Println(" No valid answer entered.  Will assume 0.")
		requestedwordcount = 0
	}

	if requestedwordcount == 0 {
		requestedwordcount = int(filesize / 7)
	}

	s := fmt.Sprintf(" filesize = %d, requestedwordcount = %d \n", filesize, requestedwordcount)
	OutBufioWriter.WriteString(s)
	mastersliceofwords := make([]string, 0, requestedwordcount)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	for totalwords := 0; totalwords < requestedwordcount; totalwords++ { // Main processing loop
		word, err := bytesbuffer.ReadString('\n')
		if err != nil {
			break
		}
		word = strings.TrimSpace(word)
		//	word = strings.ToLower(strings.TrimSpace(word))
		if len(word) < 4 {
			continue
		}
		mastersliceofwords = append(mastersliceofwords, word)
	}

	numberofwords := len(mastersliceofwords)

	allowoutput := false
	if numberofwords < 50 {
		allowoutput = true
	}

	// make the sliceofwords
	if allowoutput {
		fmt.Println("master before:", mastersliceofwords)
	}
	sliceofwords := make([]string, numberofwords)

	fmt.Println()
	fmt.Println()

	// sort.StringSlice method
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("slice before first sort.StringSlice:", sliceofwords)
	}
	NativeWords := sort.StringSlice(sliceofwords)
	t9 := time.Now()
	NativeWords.Sort()
	NativeSortTime := time.Since(t9)
	NativeSortTimeNano := NativeSortTime.Nanoseconds()
	s = fmt.Sprintf(" after sort.StringSlice: %s, %d ns \n", NativeSortTime.String(), NativeSortTimeNano)
	fmt.Println(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	//	s = fmt.Sprintf("%v\n", NativeWords)
	//	_, err = OutBufioWriter.WriteString(s)
	//	check(err)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()
	fmt.Println()
	fmt.Println()

	// TreeSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("slice before TreeSort:", sliceofwords)
	}
	t0 := time.Now()
	TreeWords := TreeSort(sliceofwords)
	TreeSortTime := time.Since(t0)
	s = fmt.Sprintf(" time after TreeSort: %s, %d ns \n", TreeSortTime.String(), TreeSortTime.Nanoseconds())
	fmt.Println(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	if allowoutput {
		for _, w := range TreeWords {
			fmt.Print(w, " ")
		}
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// Wrap it up by writing number of words, etc.
	s = fmt.Sprintf(" requestedwordcount= %d, numberofwords= %d, len(mastersliceofwords)= %d \n",
		requestedwordcount, numberofwords, len(mastersliceofwords))
	_, err = OutBufioWriter.WriteString(s)
	if len(mastersliceofwords) > 1000 {
		fmt.Println(s)
	}
	_, err = OutBufioWriter.WriteString("------------------------------------------------------\n")
	check(err)

	// Close the output file and exit
	OutBufioWriter.Flush()
	OutputFile.Close()
}

func TreeSort(input []string) []string {
	/*
	     This program creates a butterfly merge to combine leaves into twigs, twigs into branches, and branches into one final linked list starting at position LASTELEM+1.  In the end, LINKS(LASTELEM+1)
	   points to the first leaf, LINKS(LINKS(LASTELEM+1)) points to the second, LINKS(LINKS(LINKS(LASTELEM+1))) points to the third, and the last link points to itself.

	     The dimensions of LINKS needs to be >= LASTELEM + log2(LASTELEM) + 2, and the procedure will check to make sure this condition is met.  If not, a warning message is displayed on the terminal.

	*/

	/*
	     This sorting algorithm was written to use as few comparisons as possible, to have as few steps btwn each comparison as possible, to take advantage of natural sequencing, to preserve the order of
	   equals (or even the reverse order of equals), ie, to be stable, to use as little memory as possible (one working array), and to be a modular, easily understood program written in BASIC.
	   Unfortunately, the horrendous variable names in this program are from this original BASIC listing of the program.

	     The theory behind the algorithm may be described in a language of forests, trees, branches, twigs and leaves.  There is a forest filled with trees of different sizes.  Each tree is very orderly.
	   The trunk of a tree splits into two branches of nearly the same size.  If one branch is larger than the other, it is always the right-hand branch.  Similarly, each branch divides into two more
	   branches until the branches become twigs from which leaves grow.  The leaves are the individual elements to be sorted.

	     This program creates a butterfly merge to combine leaves into twigs, twigs into branches, and branches into one final linked list starting at position ELEMCNT+1.  In the end, LINKS(ELEMCNT+1) points
	   to the first leaf, LINKS(LINKS(ELEMCNT+1)) points to the second, LINKS(LINKS(LINKS(ELEMCNT+1))) points to the third, and the last link points to itself.

	     Two things happen as the pgm jumps from twig to twig.  The leaves ahead of the current record pointer get merged into a twig and the twigs and branches behind this pointer get merged into larger
	   branches.  The butterfly merge treats each merge the same way.  The heads of each twig sequence are kept at positions ELEMCNT+1, ELEMCNT+2, ..., ELEMCNT+log2(ELEMCNT)+2 after the links themselves,
	   which are kept in positions 1,2,3,...,ELEMCNT of array links.

	     The merge takes the last 2 sequences created in the list and combines them into one.  One wing of the merge follows sequence 1 and the other follows sequence 2.  The two are interwoven until the
	   final link points to itself.  Because the heads of each sequence are kept in the same array with the links themselves, the merge is extraordinarily fast.  After each merge, the stack of twig sequence
	   heads has been reduced by one.

	     Each time the current record pointer reaches a new twig, it generates new sequences one item long to correspond to the leaves of that twig.  A two leaf twig is produced by creating two one item
	   sequences, each pointing to itself.  Then these two leaves are merged once.  A three leaf twig is created from three one item sequences merged twice.  A four leaf twig is merged from two two leaf
	   twigs: the first two leaf twig is created and merged once; then the number of remaining merge passes is set to a negative number so that the merge will be disabled until the second two leaf twig is
	   created and merged with the first.

	     After each complete twig has been generated, merging continues until the branches behind the current record pointer have been linked together.  Then it jumps to the next twig, generates new leaves
	   and lets the butterfly merge fly again.

	     A state machine was the simplest way to implement all the branching of the FORTRAN code.  An arbitrary numbering from 1 to 6 is used in which the variable STATE represents this state machine's
	   indicator.  A second minor state machine had to be introduced as well, which uses the variable MODE as its indicator.  The commented out numbers are the statement labels from my FORTRAN source code
	   listing.

	*/

	var AK1, AK2, T2, T3, T4, SQNC1, B1, B2 int // I don't remember why I made these LONGREAL in the first place.
	var SEQHEAD, L0, L1, L2, STATE, ELEMIDX, T1, size int
	var MERGES int
	var Links []int

	var ItemArray []string

	AK1 = 0
	ELEMIDX = 0
	MERGES = 0
	T2 = 0
	T4 = 0
	SEQHEAD = len(input) - 1
	size = SEQHEAD + int(math.Log2(float64(len(input)))+0.5) + 2
	Links = make([]int, size) // I need max size to be len of input + log base 2 of the len of input + 2.
	Links[0] = 0              // That pesky zero origin arrays again.  Let's see if I convert correctly.
	Links[SEQHEAD] = 0
	AK2 = 1
	SQNC1 = SEQHEAD
	for SQNC1 >= 4 { // Climb the tree
		AK2 = 2 * AK2
		B2 = SQNC1 / 2
		SQNC1 = B2
		T4 += AK2 * (B2 - SQNC1)
	} // END WHILE from M-2 code

	T4 = AK2 - T4 // T4 is the # of low order twigs
	B2 = AK2 / 2
	STATE = 1

	t0 := time.Now()
StateLoop:
	for (STATE > 1) || (AK1 < AK2) { // Next twig.  Can only exit when STATE = 1
		switch STATE {
		case 1:
			AK1++
			T1 = AK1
			B1 = B2
			T3 = T2
			for !ODD(T1) {
				MERGES++
				T2 = T2 - B1
				B1 = B1 / 2
				T1 = T1 / 2
			}

			// Twig calculations
			T2 += B1
			if (SQNC1 == 2) && (T3 < T4) {
				MERGES++ // 2 twig
			} else if (SQNC1 == 2) || (T3 < T4) {
				MERGES++ // 3 twig
				ELEMIDX++

				// Make a leaf
				Links[ELEMIDX] = ELEMIDX
				Links[SEQHEAD] = ELEMIDX

				// Next sequence head
				SEQHEAD++
				MERGES++ // 2 twig
			} else { //  4 twig.  Disengage # of merges
				MERGES = -1 * MERGES
			} //END IF SQNC1 == 2

			STATE = 2

		case 2:
			ELEMIDX++

			// Make a leaf
			L1 = ELEMIDX
			Links[ELEMIDX] = ELEMIDX
			Links[SEQHEAD] = ELEMIDX

			// L0 is head of older leaf
			L0 = SEQHEAD

			// Head of most recent leaf
			SEQHEAD++
			ELEMIDX++

			// Make a leaf
			L2 = ELEMIDX
			Links[ELEMIDX] = ELEMIDX
			Links[SEQHEAD] = ELEMIDX
			STATE = 4

		case 4:
			if input[L1] <= input[L2] {
				STATE = 5 // switch to sequence 1
			} else { // switch to sequence 2
				Links[L0] = L2
				for {
					L0 = L2
					L2 = Links[L0] // next leaf
					if L2 == L0 {  // switch to sequence 1
						Links[L0] = L1
						STATE = 6
						break
					} // ENDIF
					if input[L1] <= input[L2] {
						break
					}
					elapsed := time.Since(t0)
					if elapsed > 10*time.Second {
						fmt.Println("Timeout in case 4 infinite loop")
						break StateLoop
					}
				} //END forLOOP
				if STATE /* still */ == 4 {
					Links[L0] = L1
					STATE = 5
				} //ENDIF
			} // ENDIF

		case 5:
			L0 = L1
			L1 = Links[L0]
			if L1 != L0 {
				STATE = 4
			} else {
				Links[L0] = L2
				STATE = 6
			} // ENDIF
		case 6:
			MERGES--
			if MERGES > 0 {
				SEQHEAD--           // Head of latest branch or twig
				L0 = SEQHEAD - 1    // Head of older branch or twig
				L1 = Links[L0]      // Head of sequence 1
				L2 = Links[SEQHEAD] // Head of sequence 2
				STATE = 4
			} else if MERGES == 0 {
				STATE = 1
			} else { // MERGES < 0

				// Make 2nd half of 4-twig by re-engaging the # of merges
				MERGES = -1*MERGES + 1
				STATE = 2
				elapsed := time.Since(t0)
				if elapsed > 10*time.Second {
					fmt.Println("Timeout in case 6 loop")
					break StateLoop
				}
			} // ENDIF
		default:
			elapsed := time.Since(t0)
			if elapsed > 10*time.Second {
				fmt.Println("Timeout in default case loop")
				break StateLoop
			}

		} //  END STATE MACHINE switch-case
	} //  END outer for STATE > 1 OR AK1 < AK2 loop

	fmt.Println(" Input array:", input)
	fmt.Println(" Links Array:", Links)
	fmt.Println("len(input)=", len(input))
	fmt.Println()


	ItemArray = make([]string, 0, len(input))

	k := Links[len(input)]

	//  for k == LINKS[k] may be alternative approach, if I cannot define i but not use it.
	for i := 0; i < len(input); i++ {
		ItemArray = append(ItemArray, input[Links[k]])
		k = Links[k]
	}
	return ItemArray
} //  END TREESORT;

func ODD(i int) bool {
	return i%2 == 1
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
{{{
MODULE FREQ;

(*
  REVISION HISTORY
  ----------------
*)
  FROM SYSTEM IMPORT ADR,DOSCALL
(*Error  : Unsupported SYSTEM item ==> 'DOSCALL' *);
  IMPORT Break;
  IMPORT DebugPMD;
  FROM Terminal IMPORT ReadString;
  IMPORT Terminal;
  FROM InOut IMPORT ReadCard,Read,WriteString,WriteLn,WriteCard,WriteInt,
    WriteHex,WriteOct,Write;
  FROM LongIO IMPORT ReadLongInt, WriteLongInt;
  FROM FloatingUtilities IMPORT Frac,Int,Round,Float,Trunc;
  FROM UTILLIB IMPORT MAXCARD,CR,LF,NULL,BUFSIZ,CTRLCOD,STRTYP,STR10TYP,
    BUFTYP,MAXCARDFNT;
  FROM UL2 IMPORT COPYLEFT,COPYRIGHT,FILLCHAR,SCANFWD,SCANBACK,CopyWords,
    FillWord,STRLENFNT,STRCMPFNT,LCHINBUFFNT,MRGBUFS,TRIMFNT,TRIM,RMVCHR,
    APPENDA2B,CONCATAB2C,INSERTAin2B,ASSIGN2BUF,GETFNM;
  FROM FIO IMPORT DRIVESEP,SUBDIRSEP,MYFILTYP,IOSTATE,FRESET,FCLOSE,
    FRDTXLN,FWRTXLN,FWRTX,RETBLKBUF,EXTRACTDRVPTH,FCLOSEDUP,FWRLN;
  FROM TOKENIZE IMPORT FSATYP,DELIMCH,INI1TKN,GETCHR,UNGETCHR,GETTKN,
    GETTKNREAL,UNGETTKN;
(*  FROM SORTER IMPORT ITEMARRAY,LINKS,TREESORT,QUICKSORT,BININSSORT;*)


  CONST
    MAXLONGINT = 7FFFFFFFH;
    SP = ORD(' ');
    ORDAT = ORD('@');
    ORDCAPA = ORD('A');
    ORDCAPZ = ORD('Z');
    N  = 26;

TYPE
  INDEX = INTEGER;
  ITEM  = LONGINT;

  VAR
    LNCTR,C,ORDCAPCH                                : CARDINAL;
    INUN1                                           : MYFILTYP;
    PROMPT,NAMDFT,TYPDFT,BLANKBUF,INFNAM,INPUT,BUF  : BUFTYP;
    EOFFLG                                          : BOOLEAN;
  LETTERTABLE : ARRAY [1..N]  OF ITEM;
  LINKS     : ARRAY [1..2*N]  OF INDEX;
    REVLINKS                        : ARRAY [1..N] OF CARDINAL;

PROCEDURE TREESORT(LASTELEM : CARDINAL);

  VAR
    AK1,AK2,T2,T3,T4,SQNC1,      B1,B2     : LONGREAL;
    SEQHEAD,L0,L1,L2,STATE,MODE,ELEMIDX,T1 : CARDINAL;
    MERGES                                 : INTEGER;

  BEGIN
    AK1 = 0.;
    ELEMIDX = 0;
    MERGES = 0;
    T2 = 0.;
    T4 = 0.;
    SEQHEAD = LASTELEM + 1;
    LINKS[1] = 1;       (* That pesky zero origin arrays again *)
    LINKS[SEQHEAD] = 1;
    AK2 = 1.;
    SQNC1 = FLOAT(LASTELEM);
    WHILE SQNC1 >= 4. DO    (* Climb the tree *)
      AK2 = 2.*AK2;
      B2 = SQNC1/2.;
      SQNC1 = Int(B2);
      T4 = T4 + AK2*(B2-SQNC1);
    END(*WHILE*);
    T4 = AK2 - T4; (* T4 is the # of low order twigs *)
    B2 = AK2/2.;
    STATE = 1;
(* 13 *)
    WHILE (STATE > 1) OR (AK1 < AK2) DO (* Next twig.  Can only exit when *)
      CASE STATE OF                     (* STATE = 1 *)
      1:AK1 = AK1 + 1.;
        T1 = Round(AK1);
        B1 = B2;
        T3 = T2;
        WHILE NOT ODD(T1) DO
          INC(MERGES);
          T2 = T2 - B1;
          B1 = B1/2.;
          T1 = T1 DIV 2;
        END(*LOOP*);
(* Twig calculations *)
        T2 = T2 + B1;
        IF (SQNC1 = 2.) AND (T3 < T4) THEN
          INC(MERGES);  (* 2 twig *)
        ELSIF (SQNC1 = 2.) OR (T3 < T4) THEN
(* 5 *)   INC(MERGES);     (* 3 twig *)
          INC(ELEMIDX);
(* Make a leaf *)
          LINKS[ELEMIDX] = ELEMIDX;
          LINKS[SEQHEAD] = ELEMIDX;
(* Next sequence head *)
          INC(SEQHEAD);
          INC(MERGES);  (* 2 twig *)
        ELSE  (* 4 twig.  Disengage # of merges *)
          MERGES = -1*MERGES;
        END(*IF*);
        STATE = 2;
      | 2:
(* 7 *) INC(ELEMIDX);
(* Make a leaf *)
        L1 = ELEMIDX;
        LINKS[ELEMIDX] = ELEMIDX;
        LINKS[SEQHEAD] = ELEMIDX;
(* L0 is head of older leaf *)
        L0 = SEQHEAD;
(* Head of most recent leaf *)
        INC(SEQHEAD);
        INC(ELEMIDX);
(* Make a leaf *)
        L2 = ELEMIDX;
        LINKS[ELEMIDX] = ELEMIDX;
        LINKS[SEQHEAD] = ELEMIDX;
        STATE = 4;
      | 4:
(* 9 *) IF LETTERTABLE[L1] <= LETTERTABLE[L2] THEN
          STATE = 5 (* switch to sequence 1 *)
        ELSE (* switch to sequence 2 *)
          LINKS[L0] = L2;
(* 8 *)   LOOP
            L0 = L2;
            L2 = LINKS[L0]; (* next leaf *)
            IF L2 = L0 THEN (* switch to sequence 1 *)
              LINKS[L0] = L1;
              STATE = 6;
              EXIT;
            END(*IF*);
            IF (LETTERTABLE[L1] <= LETTERTABLE[L2]) THEN EXIT END(*IF*);
          END(*LOOP*);
          IF STATE (* still *) = 4 THEN
            LINKS[L0] = L1;
            STATE = 5;
          END(*IF*);
        END(*IF*);
      | 5:
(* 11 *)L0 = L1;
        L1 = LINKS[L0];
        IF L1 != L0 THEN
          STATE = 4;
        ELSE
          LINKS[L0] = L2;
          STATE = 6;
        END(*IF*);
      | 6:
(* 10 *)DEC(MERGES);
        IF MERGES > 0 THEN
(* 12 *)  DEC(SEQHEAD);  (* Head of latest branch or twig *)
          L0 = SEQHEAD - 1;    (* Head of older branch or twig *)
          L1 = LINKS[L0];      (* Head of sequence 1 *)
          L2 = LINKS[SEQHEAD]; (* Head of sequence 2 *)
          STATE = 4;
        ELSIF MERGES = 0 THEN
          STATE = 1
        ELSE (* MERGES < 0 *)
(* Make 2nd half of 4-twig by re-engaging the # of merges *)
          MERGES = -1*MERGES + 1;
          STATE = 2;
        END(*IF*);
      END(*STATE MACHINE CASE*);
    END(*WHILE*);
  END TREESORT;

  PROCEDURE PRINTSORTED;

  VAR
    C,I,J,K : CARDINAL;

  BEGIN
    TREESORT(N);

    WriteString(' TreeSorted Array: ');
    WriteLn;
    K = LINKS[N+1];
    FOR C = 1 TO N DO   I reversed this so I can print the highest first instead of the lowest.  At the time, that was easier than reversing the algorithm.
      REVLINKS[C] := K;
      K = LINKS[K];
    END(*FOR*);
    FOR C = N TO 1 BY -1 DO
      Write(CHR(REVLINKS[C]+ORDAT));
      K = LINKS[K];
    END(*FOR*);
    WriteLn;

  END PRINTSORTED;

  BEGIN
    ASSIGN2BUF(' Enter Input File Name : ',PROMPT);
    ASSIGN2BUF('',NAMDFT);
    ASSIGN2BUF('.DOC',TYPDFT);
    GETFNM(PROMPT, NAMDFT, TYPDFT, INFNAM);
    WriteString(' Input File Name : ');
    WriteString(INFNAM.CHARS);
    WriteLn;
    FRESET(INUN1,INFNAM,RD);

    LNCTR = 0;
    FOR C = 1 TO N DO LETTERTABLE[C] = 0; END(*FOR*);

    LOOP
      FRDTXLN(INUN1,BUF,EOFFLG);
      IF EOFFLG THEN EXIT END(*IF*);
      INC(LNCTR);
      C = 1;
      WHILE C <= BUF.COUNT DO     (* Will skip blank lines *)
        ORDCAPCH = ORD(CAP(BUF.CHARS[C]));
        INC(C);
        IF (ORDCAPCH >= ORDCAPA) AND (ORDCAPCH <= ORDCAPZ) THEN
          DEC(ORDCAPCH,ORDAT);
          LETTERTABLE[ORDCAPCH] = LETTERTABLE[ORDCAPCH] + 1;
          IF LETTERTABLE[ORDCAPCH] = MAXLONGINT THEN EXIT; END(*IF*);
        END(*IF*);
      END(*WHILE*);
    END(*LOOP*);

    FCLOSE(INUN1);

    WriteString(' Processed ');
    WriteCard(LNCTR,0);
    WriteString(' lines.');
    WriteLn;

    FOR C = 1 TO N DO
      Write(' ');
      Write(CHR(C+ORDAT));
      WriteString(' = ');
      WriteLongInt(LETTERTABLE[C],10);
      IF C MOD 5 = 0 THEN WriteLn END(*IF*);
    END(*FOR*);
    WriteLn;

    PRINTSORTED;

  END FREQ.

func round64(f float64) float64 {
  return math.Floor(f + 0.5)
}

func float64equal(x,y float64) bool {
  return x - y < floatequaltolerance
}
}}}
*/
