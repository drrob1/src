package tknptr

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

/*
 Copyright (C) 1987-2023  Robert Solomon MD.  All rights reserved.
 These routines collectively implement a very good facility to fetch, manipulate, and interpret tokens.

REVISION HISTORY
----------------
28 MAY 87 -- Added UNGETTKN capability and no longer exported GETCHR and UNGETCHR.
29 AUG 87 -- Restored exportation of GETCHR and UNGETCHR.
 3 Mar 88 -- Added the ASCZERO declaration and removed the function call from the DGT conversion loop.
31 Mar 88 -- Converted to M2 V3.03.
 1 Sept 88 -- 1.  Allowed quoted string to force ALLELSE state.
              2.  Changed the method UNGETTKN uses to unget the token.
              3.  Added the MULTSIGN and DIVSIGN OP states.
              4.  Ran M2CHECK and deleted all unreferenced var's.
              5.  Moved the NEGATV check for contigently making SUM < 0 out
                   of the LOOP and deleted the 5 previous statements for all
                   of the different states plus the end-of-line test.
18 Mar 89 -- Added the GETTKNREAL Procedure.
20 Mar 89 -- Changed GETOPCODE so that if a multicharacter op code is invalid, UNGETCHR is used to put the second char back.
 1 Dec 89 -- Made change in GETTKN that was demonstrated to be necessary when the code was ported to the VAX.
 2 Jan 90 -- Changed GETTKNREAL so that a real number may begin with a decimal pt.
 9 Nov 90 -- Added GETTKNSTR procedure and DELIMSTATE var.
27 Dec 90 -- Added GETTKNEOL procedure, originally written for CFNTS.
25 Jul 93 -- Fixed bug in GETCHR whereby CHRSTATE not set when at EOL, and adjusted algorithm of GETTKNSTR.
 6 Jun 95 -- Added FSAARRAY as way to assign FSATYP, and to easily modify the FSATYP assignments.
20 Oct 02 -- Converted to M2 for win32, DOS mode.
17 May 03 -- First Win32 version.
30 Jun 03 -- Fixed real tokens so can now again begin w/ decpt, by always writing a leading 0.
21 Jul 03 -- Fixed bug introduced by above step when a token has leading spaces
 4 Oct 03 -- Fixed bug when neg number is entered using unary minus.
 9 Oct 13 -- Converted to gm2.
11 Oct 13 -- Fixed a bug in GETTKNREAL in which number like 1e-1 lost the e.
12 Oct 13 -- Removed an errant RETURN from GETTKNSTR.
 3 Feb 14 -- Converted to Ada.  I modernized the data types to be a record type.
28 Jun 14 -- Backported enhancement to GetOpCode that includes ^, ** and %.
19 Nov 14 -- Converted to C++.
 7 Dec 14 -- Removed comma as a delim, making it AllElse so it works as intended for HPCALCC
28 Dec 14 -- Turns out that CentOS C++ does not support -std=c++11, so I have to remove string.front and string.back member functions.
18 Jan 15 -- Found bug in that single digits followed by add or subtract are not processed correctly by GETTKNREAL.
19 Aug 16 -- Finished conversion to Go, started 8/6/16 on boat to Bermuda.
21 Sep 16 -- Now that this code is for case sensitive filesystem like linux, returning an all caps token is a bad idea.
               So I added FetchToken which takes a param of true for cap and false for preserving case.
 9 Oct 16 -- Will allow "0x" as prefix for hex, as well as "H" suffix.  An 'x' anywhere in the number will
                be a hex number.  I will not force it to be the 2nd character.
25 Nov 16 -- The TKNMAXSIZ was too small for sha512, so I increased it.
 3 Dec 16 -- Decided to change how the UpperCase flag is handled in GetToken.
10 Aug 17 -- Making this use pointer receivers, if I can.
13 Oct 17 -- Made tab char a delim.  Needed for comparehashes.
18 Oct 17 -- Changed init process so all control codes are delims, just as in the current tokenize.
19 Oct 17 -- Standard hash256 files for linux include a * in front of the filename.  I'm not sure why.  I want to
                 ignore this, so I'm writing SetMapDelim so I can.
27 Jan 18 -- Turns out that SetMapDelim doesn't work on GetTokenString, so I have to be more selective
                 when I remap the characters.
28 Sep 20 -- Now that I'm using tknptr in comparehashes, I'm going to include the statemap in the bufferstate structure
                 so it's not global.
23 Oct 20 -- GetOpcode will unget characters as needed to keep length of opcode token to a max of 2 characters.
 6 Jun 21 -- Writing GetTokenSlice, meaning return a slice of all tokens on the line, using GetToken to fetch them.
               And added a check against an empty string being passed into the init functions.
12 Jun 21 -- Writing TokenRealSlice, and renamed GetTokenSlice to TokenSlice, which is more idiomatic for Go.
 1 Apr 23 -- Added New function to not return a nil pointer if the entered string is empty.  So far, it seems to be working (tested by gonumsolve).
               I intend for this to be the preferred way to initialize a token pointer.
 4 Jul 23 -- I'm adding a much simpler way to get a real token, TokenReal.  This is going to take a while because I'm also listening to Derek Parker's 8 days of debugging.
              And I added '_' following 'E' or 'e' is replaced by '-' before conversion to float.
 9 Jul 23 -- Now to fix hex input, by adding a field to TokenType.
19 Jul 23 -- Added TokenRealSlice, which uses the new TokenReal(), instead of the old GETTKNREAL().  And I decided to not allow '-' for negative exponents.  I very rarely use E notation
               so I want to preserve the use of '-' as an operator.  Underscore, '_', is replaced w/ '-' just before call to strconv.ParseFloat().
*/

const LastAltered = "22 July 2023"

const (
	DELIM = iota // so DELIM = 0, and so on.  And the zero val needs to be DELIM.
	OP
	DGT
	ALLELSE
)

type TokenType struct {
	Str        string
	FullString string // includes minus sign character, if present.
	State      int
	DelimCH    byte
	DelimState int
	Isum       int
	Rsum       float64
	RealFlag   bool // flag so integer processing stops when it sees a dot, E or e.  And to use strconv.ParseFloat for the conversion.
	HexFlag    bool // only way I know of to signal that the input string is a hex format.
} // TokenType record

type CharType struct {
	Ch    byte
	State int
} // CharType Record

type BufferState struct {
	CURPOSN, HOLDCURPOSN, PREVPOSN int
	lineByteSlice, HoldLineBS      []byte
	StateMap                       map[byte]int // as of 9/28/20, StateMap is part of this structure.
}

// ---------------------------------------------------------------------

const TKNMAXSIZ = 180
const OpMaxSize = 2
const Dgt0 = '0'
const Dgt9 = '9'
const POUNDSIGN = '#' // 35
const PLUSSIGN = '+'  // 43
const COMMA = ','     // 44
const MINUSSIGN = '-' // 45
const SEMCOL = ';'    // 59
const LTSIGN = '<'    // 60
const EQUALSIGN = '=' // 61
const GTSIGN = '>'    // 62
const MULTSIGN = '*'
const DIVSIGN = '/'
const SQUOTE = '\047' // 39
const DQUOTE = '"'    // 34
const NullChar = 0
const EXPSIGN = '^'
const PERCNT = '%'

var wantReal bool // used by TokenReal

// These variables are declared here to make the variable global so to maintain their values btwn calls.

func Cap(c rune) rune {
	r, _, _, _ := strconv.UnquoteChar(strings.ToUpper(string(c)), 0)
	return r
} // Cap

/* strconv
  func UnquoteChar

func UnquoteChar(s string, quote byte) (value rune, multibyte bool, tail string, err error)

UnquoteChar decodes the first character or byte in the escaped string or character literal represented by the string s. It returns four values:

1) value, the decoded Unicode code point or byte value;
2) multibyte, a boolean indicating whether the decoded character requires a multibyte UTF-8 representation;
3) tail, the remainder of the string after the character; and
4) an error that will be nil if the character is syntactically valid.

The second argument, quote, specifies the type of literal being parsed and therefore which escaped quote character is permitted. If set to a single quote,
it permits the sequence \' and disallows unescaped '. If set to a double quote, it permits \" and disallows unescaped ". If set to zero, it does not permit
either escape and allows both quote characters to appear unescaped.
*/

func CAP(c byte) byte {
	if (c >= 'a') && (c <= 'z') {
		c = c - 32
	}
	return c
}

// ----------------------------------------- init --------------------------------------------
/* removed 9/28/20 when StateMap became part of BufferState
func init() {
	StateMap = make(map[byte]int, 128)
	InitStateMap()
}

*/

// ------------------------------------ InitStateMap ------------------------------------------------
// Making sure that the StateMap is at its default values, since a call to GetTokenStr changes some values.

func InitStateMap(bs *BufferState) {
	bs.StateMap = make(map[byte]int, 128)
	//	StateMap[NullChar] = DELIM
	for i := 0; i < 33; i++ { // note that this includes \t, HT, tab character.
		bs.StateMap[byte(i)] = DELIM
	}
	for i := 33; i < 128; i++ {
		bs.StateMap[byte(i)] = ALLELSE // including comma
	}
	for c := Dgt0; c <= Dgt9; c++ {
		bs.StateMap[byte(c)] = DGT
	}
	bs.StateMap[' '] = DELIM
	bs.StateMap[';'] = DELIM
	bs.StateMap['#'] = OP
	bs.StateMap['*'] = OP
	bs.StateMap['+'] = OP
	bs.StateMap['-'] = OP
	bs.StateMap['/'] = OP
	bs.StateMap['<'] = OP
	bs.StateMap['='] = OP
	bs.StateMap['>'] = OP
	bs.StateMap['%'] = OP
	bs.StateMap['^'] = OP
} // InitStateMap

// ----------------------------------------- INITKN -------------------------------------------

func INITKN(Str string) *BufferState { // constructor, initializer
	// INITIALIZE TOKEN.
	// The purpose of the initialize token routine is to initialize the
	// variables used by nxtchr to begin processing a new line.
	// The buffer on which the tokenizing rtns operate is also initialized.
	// CURPOSN is initialized to start at the first character on the line.
	if Str == "" {
		return nil
	}

	bs := new(BufferState) // idiomatic Go would write this as &BufferState{}
	InitStateMap(bs)       // It's possible GetTknStr or GetTknEOL changed the StateMap, so will call init.
	bs.CURPOSN, bs.PREVPOSN, bs.HOLDCURPOSN = 0, 0, 0
	bs.lineByteSlice = []byte(Str)
	copy(bs.HoldLineBS, bs.lineByteSlice) // make sure that a value is copied, not just a pointer.
	/*
	   fmt.Println(" In IniTkn and Str is:",Str,", len of Str is: ",len(Str));
	   fmt.Print(" In IniTkn and linebyteslice is '",bs.lineByteSlice);
	   fmt.Println("', length of lineByteSlice is ",len(bs.lineByteSlice));
	   fmt.Println(" In IniTkn and Str is '",Str,"', length= ",len(Str));
	*/
	return bs
} // INITKN

// ----------------------------------------- NewToken ----------------------------------------
/*
func NewToken(Str string) *BufferState { // constructor, initializer
	// INITIALIZE TOKEN, using the Go idiom.

	if Str == "" {
		return nil
	}
	bs := new(BufferState) // idiomatic Go would write this as &BufferState{}
	InitStateMap(bs)       // possible that GetTknStr or GetTknEOL changed the StateMap, so will call init.
	bs.CURPOSN, bs.PREVPOSN, bs.HOLDCURPOSN = 0, 0, 0
	bs.lineByteSlice = []byte(Str)
	copy(bs.HoldLineBS, bs.lineByteSlice) // make sure that a value is copied.
	return bs
} // NewToken, copied from INITKN
*/

// ----------------------------------------- New ----------------------------------------

func New(Str string) *BufferState { // constructor, initializer
	// INITIALIZE TOKEN, using the Go idiom.

	bs := new(BufferState) // idiomatic Go would write this as &BufferState{}
	//bs = &BufferState{}
	InitStateMap(bs) // possible that GetTknStr or GetTknEOL changed the StateMap, so will call init.
	bs.CURPOSN, bs.PREVPOSN, bs.HOLDCURPOSN = 0, 0, 0
	bs.lineByteSlice = []byte(Str)
	copy(bs.HoldLineBS, bs.lineByteSlice) // make sure that a value is copied.
	return bs
} // New, copied from NewToken

//------------------------------ STOTKNPOSN -----------------------------------

func (bs *BufferState) STOTKNPOSN() {
	// STORE TOKEN POSITION.
	// This routine will store the value of the curposn into a hold variable for later recall by RCLTKNPOSN.

	if (bs.CURPOSN < 0) || (bs.CURPOSN > len(bs.lineByteSlice)) {
		log.SetFlags(log.Llongfile)
		log.Print(" In StoTknPosn and CurPosn is invalid.")
		os.Exit(1)
	}
	bs.HOLDCURPOSN = bs.CURPOSN
	copy(bs.HoldLineBS, bs.lineByteSlice) // Need to copy values, else just copy a pointer
} // STOTKNPOSN

//------------------------------ RCLTKNPOSN ----------------------------------

func (bs *BufferState) RCLTKNPOSN() {
	/*
	   RECALL TOKEN POSITION.
	   THIS IS THE INVERSE OF THE STOTKNPOSN PROCEDURE.
	*/

	if (bs.HOLDCURPOSN < 0) || (len(bs.HoldLineBS) == 0) || (bs.HOLDCURPOSN > len(bs.HoldLineBS)) {
		log.SetFlags(log.Llongfile)
		log.Print(" In RclTknPosn and HoldCurPosn is invalid.")
	}
	bs.CURPOSN = bs.HOLDCURPOSN
} // RCLTKNPOSN

// ---------------------------- PeekChr -------------------------------------------

func (bs *BufferState) PeekChr() (CharType, bool) {
	/*
	   -- This is the GET CHARACTER ROUTINE.  Its purpose is to get the next
	   -- character from inbuf, determine its fsatyp (finite state automata type),
	   -- and return the upper case value of char.
	   -- NOTE: the curposn pointer is used before it's incremented, unlike most of the
	   -- other pointers in this program.
	      As of 21 Sep 16, the CAP function was removed from here, and conditionally placed into GetToken.
	*/
	var C CharType
	var EOL bool
	EOL = false
	if bs.CURPOSN >= len(bs.lineByteSlice) {
		EOL = true
		//C = CharType{} // This zeros the CharType C by assigning an empty CharType constant literal.  Not necessary, as the var declaration does the same thing.
		return C, EOL
	}
	C.Ch = bs.lineByteSlice[bs.CURPOSN] //  no longer use the Cap function here.
	C.State = bs.StateMap[C.Ch]         // state assignment, here using map access.
	return C, EOL
} // PeekCHR

// ---------------------------- NextChr  -------------------------------------------

func (bs *BufferState) NextChr() {
	bs.CURPOSN++
} // NextChr

// --------------------------------- GetChr --------------------------------

func (bs *BufferState) GETCHR() (CharType, bool) {
	C, EOL := bs.PeekChr()
	bs.NextChr()
	return C, EOL
} // GETCHR

// --------------------------------- UNGETCHR --------------------------------

func (bs *BufferState) UNGETCHR() {
	/*
	   -- UNGETCHaracteR.
	   -- This is the routine that will allow the character last read to be read
	   -- again by decrementing the pointer into the line buffer, CURPOSN.
	*/

	if bs.CURPOSN < 0 {
		log.SetFlags(log.Llongfile)
		log.Print(" CURPOSN out of range in UnGetChr")
		os.Exit(1)
	}
	bs.CURPOSN--
} // UNGETCHR

// ------------------------------------- GetOpCode ---------------------------------------------
// I am not coding this as a pointer receiver as it does not access its BufferState.

func (bs *BufferState) GETOPCODE(Token TokenType) int {
	/*
	   -- GET OPCODE.
	   -- This routine receives a token of FSATYP op (meaning it is an operator)
	   -- and analyzes it to determine an opcode, which is a nUMBER from 1..22.
	   -- This is done after the necessary validity check of the input token.
	   -- The opcode assignments for the op tokens are:
	   --  < is 1                  <= is 2
	   --  > is 3                  >= is 4
	   --  = is 5   == is 5        <> is 6    # is 7
	   --  + is 8                  += is 9
	   --  - is 10                 -= is 11
	   --  * is 12                 *= is 13
	   --  / is 14                 /= is 15
	   --  ^ is 16                 ^= is 17
	   -- ** is 18                **= is too long to be allowed
	   -- >< is 20
	   --  % is 22
	*/
	var CH1, CH2 byte
	OpCode := 0

	if len(Token.Str) < 1 {
		log.SetFlags(log.Llongfile)
		log.Print(" Token is empty, from GetOpCode.")
		return OpCode
	}

	// keep token length to 2 chars max.
	for len(Token.Str) > 2 {
		bs.UNGETCHR()
		Token.Str = Token.Str[:len(Token.Str)-1]
	}

	CH1 = Token.Str[0]
	if len(Token.Str) > 1 {
		CH2 = Token.Str[1]
	} else {
		CH2 = NullChar
	}

	switch CH1 {
	case LTSIGN:
		OpCode = 1
	case GTSIGN:
		OpCode = 3
	case EQUALSIGN:
		OpCode = 5
	case PLUSSIGN:
		OpCode = 8
	case MINUSSIGN:
		OpCode = 10
	case POUNDSIGN:
		OpCode = 7
	case MULTSIGN:
		OpCode = 12
	case DIVSIGN:
		OpCode = 14
	case EXPSIGN:
		OpCode = 16
	case PERCNT:
		OpCode = 22
	default:
		return OpCode
	} // switch case

	if len(Token.Str) == 1 {
		return OpCode
	} else if (CH2 == EQUALSIGN) && (CH1 != EQUALSIGN) && (CH1 != POUNDSIGN) {
		OpCode++
	} else if (CH1 == LTSIGN) && (CH2 == GTSIGN) {
		OpCode = 6
	} else if (CH1 == GTSIGN) && (CH2 == LTSIGN) {
		OpCode = 20
	} else if (CH1 == MULTSIGN) && (CH2 == MULTSIGN) {
		OpCode = 18
	} else if (CH1 == EQUALSIGN) && (CH2 == EQUALSIGN) {
		// do nothing
	} else { // have invalid pair, like +- or =>.
		bs.UNGETCHR() // unget the 2nd part of the invalid pair.
	} // Length of Token = 1
	return OpCode
} // GETOPCODE

//       ---------------------------=== GetToken ===--------------------------------------

func (bs *BufferState) GetToken(UpperCase bool) (TOKEN TokenType, EOL bool) {

	var CHAR CharType
	var QUOCHR rune // Holds the active quote char

	bs.PREVPOSN = bs.CURPOSN
	var NEGATV, QUOFLG bool
	TOKEN = TokenType{}                    // This will zero out all the fields by using a nil struct literal.  It's the default; I put it here so I remember.
	tokenByteSlice := make([]byte, 0, 200) // to build up the TOKEN.Str field

ExitForLoop:
	for {
		CHAR, EOL = bs.GETCHR()
		if EOL {
			//          If TKNSTATE is DELIM, then gettkn was called when there were
			//          no more tokens on line.  Otherwise it means that we have fetched the last
			//          TOKEN on this line.
			if (TOKEN.State == DELIM) && (len(tokenByteSlice) == 0) {
				break // with EOL still being true.
			} else { // now on last token of line, so don't return with EOL set to true.
				EOL = false
			}
		} // if EOL
		if QUOFLG && (CHAR.Ch != NullChar) {
			CHAR.State = ALLELSE
		}

		switch TOKEN.State {
		case DELIM: // token.state
			switch CHAR.State {
			case DELIM: // NULL char is a special delimiter because it will
				// immediately cause a return even if there is no token yet,
				// i.e., the token is only delimiters.  This is because of
				// the NULL char is the string terminater for general strings
				// and especially for environment strings, for which this
				// TOKENIZE module was originally written.
				if CHAR.Ch == NullChar {
					break ExitForLoop
				} // goto ExitLoop }
			case OP: // Delim -> OP means this is the 1st char of the operator token.
				TOKEN.State = OP
				tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
				if len(tokenByteSlice) > TKNMAXSIZ {
					log.SetFlags(log.Llongfile)
					log.Println(" token too long in GetToken.")
					os.Exit(1)
				}
			case DGT: // Delim -> DGT means this is the 1st dgt for the entered number.
				tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
				TOKEN.State = DGT
				//bs.StateMap['X'] = DGT
				//bs.StateMap['x'] = DGT
				//bs.StateMap['H'] = DGT
				//bs.StateMap['h'] = DGT
				if wantReal {
					bs.StateMap['E'] = DGT
					bs.StateMap['e'] = DGT
				}
				if !unicode.IsDigit(rune(CHAR.Ch)) {
					TOKEN.RealFlag = true
					continue
				}
				TOKEN.Isum = int(CHAR.Ch) - Dgt0
			case ALLELSE: // Delim -> AllElse means this is first char of this alphanumeric token.
				TOKEN.State = ALLELSE
				QUOFLG = (CHAR.Ch == SQUOTE) || (CHAR.Ch == DQUOTE)
				if QUOFLG { // Do not put the quote character into the token.
					QUOCHR = rune(CHAR.Ch)
				} else {
					tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
					if len(tokenByteSlice) > TKNMAXSIZ {
						log.SetFlags(log.Llongfile)
						log.Println(" token too long in GetToken.")
						os.Exit(1)
					} // if token too long
					TOKEN.Isum = int(CHAR.Ch)
				} // QUOFLG
			} // Of Char.State
		case OP: // token.state
			switch CHAR.State {
			case DELIM:
				bs.UNGETCHR()     // To allow correct processing of op pair that is not a valid op, like +- or =>
				break ExitForLoop //goto ExitLoop;
			case OP: // OP -> OP means another operator character found.
				if len(tokenByteSlice) > OpMaxSize {
					bs.UNGETCHR()
					break ExitForLoop //goto ExitLoop;
				}
				tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
			case DGT: // OP -> DGT means it may be a sign character for a number token.  If not, have 1 char operator
				upperbound := len(tokenByteSlice) - 1
				LastChar := tokenByteSlice[upperbound]
				if (LastChar == '+') || (LastChar == '-') {
					if len(tokenByteSlice) == 1 {
						if tokenByteSlice[0] == '-' {
							NEGATV = true
						}
						TOKEN.State = DGT
						tokenByteSlice[0] = CHAR.Ch // OVERWRITE ARITHMETIC SIGN CHARACTER
						if wantReal {
							bs.StateMap['E'] = DGT
							bs.StateMap['e'] = DGT
						}
						if !unicode.IsDigit(rune(CHAR.Ch)) {
							TOKEN.RealFlag = true
							continue
						}
						TOKEN.Isum = int(CHAR.Ch) - Dgt0
					} else { // TOKEN length > 1 so must first return valid OP
						bs.UNGETCHR()                                // UNGET THIS DIGIT CHAR
						bs.UNGETCHR()                                // THEN UNGET THE ARITH SIGN CHAR
						CHAR.Ch = LastChar                           // SO DELIMCH CORRECTLY RETURNS THE ARITH SIGN CHAR
						tokenByteSlice = tokenByteSlice[:upperbound] // recall that upperbound is excluded in this syntax.
						//                  TOKEN.Str = TOKEN.Str[:upperbound]; // del last char of the token which is the sign character
						break ExitForLoop //goto ExitLoop;
					} // if length of the token = 1
				} else { // IF have a sign character as the lastchar
					bs.UNGETCHR()
					break ExitForLoop //goto ExitLoop;
				} // If have a sign character as the lastchar
			case ALLELSE: // OP -> AllElse
				bs.UNGETCHR()
				break ExitForLoop //goto ExitLoop;
			} // Char.State
		case DGT: // tokenstate
			switch CHAR.State {
			case DELIM:
				bs.StateMap['_'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				bs.StateMap['.'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				bs.StateMap['H'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				bs.StateMap['E'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				bs.StateMap['X'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				bs.StateMap['h'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				bs.StateMap['e'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				bs.StateMap['x'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				break ExitForLoop
			case OP: // DGT -> OP
				bs.UNGETCHR()
				bs.StateMap['_'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				break ExitForLoop          //goto ExitLoop;
			case DGT: // DGT -> DGT so we have another digit.
				tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
				if CAP(CHAR.Ch) == 'E' {
					bs.StateMap['_'] = DGT // make the underscore to be of DGT type so it will be allowed in the number
					//bs.StateMap['-'] = DGT // make the minus sign to be of DGT type so it will be allowed in the number.  Nope, changed my mind.  I almost never use E notation.
				}
				if TOKEN.RealFlag { // Isum only will contain the int part of a float.
					continue
				}
				if wantReal {
					bs.StateMap['E'] = DGT
					bs.StateMap['e'] = DGT
				}
				if !unicode.IsDigit(rune(CHAR.Ch)) {
					TOKEN.RealFlag = true
					continue
				}
				TOKEN.Isum = 10*TOKEN.Isum + int(CHAR.Ch) - Dgt0 // this total will not be correct when floating point chars, like dot and 'E' or 'e', are input.
			case ALLELSE: // DGT -> AllElse
				if rune(CHAR.Ch) == 'x' || rune(CHAR.Ch) == 'X' || rune(CHAR.Ch) == 'h' || rune(CHAR.Ch) == 'H' { // this logic isn't really correct, as it will allow 0h and terminating x to mean hex.
					TOKEN.HexFlag = true
					continue
				}

				bs.UNGETCHR()
				break ExitForLoop //goto ExitLoop;
			} // Char.State
		case ALLELSE: // tokenstate
			switch CHAR.State {
			case DELIM:
				//  Always exit if get a NULL char as a delim.  A quoted string can only get here if CH is NULL.
				break ExitForLoop //goto ExitLoop;
			case OP:
				bs.UNGETCHR()
				break ExitForLoop //goto ExitLoop;
			case DGT: // AllElse -> DGT means have alphanumeric token.
				if len(tokenByteSlice) > TKNMAXSIZ {
					log.SetFlags(log.Llongfile)
					log.Println(" token too long in GetTkn AllElse to Digit branch.")
					os.Exit(1)
				} // if token too long
				tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
				TOKEN.Isum += int(CHAR.Ch)
			case ALLELSE: // AllElse -> AllElse
				if rune(CHAR.Ch) == QUOCHR {
					QUOFLG = false
					CHAR.State = DELIM // So that DELIMSTATE will = delim
					break ExitForLoop

				} else {
					if len(tokenByteSlice) > TKNMAXSIZ {
						log.SetFlags(log.Llongfile)
						log.Println(" token too long in GetTkn AllElse -> AllElse branch.")
					} // if token too long
					tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
					TOKEN.Isum += int(CHAR.Ch)
				} // if char is a quote char
			} // Char.State
		} // Token.State
	} //LOOP to process characters

	if UpperCase {
		TOKEN.Str = strings.ToUpper(string(tokenByteSlice))
	} else {
		TOKEN.Str = string(tokenByteSlice) // Trying to apply idiomatic Go guidelines to use byte slice intermediate.
	}
	TOKEN.DelimCH = CHAR.Ch
	TOKEN.DelimState = CHAR.State
	TOKEN.FullString = TOKEN.Str
	if TOKEN.RealFlag {
		TOKEN.Str = strings.ReplaceAll(TOKEN.Str, "_", "-") // Note that '_' could mean '-' for exponents.
	}
	if TOKEN.State == DGT {
		bs.StateMap['_'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
		bs.StateMap['-'] = OP      // make sure the minus sign is back to the type it's supposed to be.
		bs.StateMap['H'] = ALLELSE
		bs.StateMap['E'] = ALLELSE
		bs.StateMap['X'] = ALLELSE
		bs.StateMap['h'] = ALLELSE
		bs.StateMap['e'] = ALLELSE
		bs.StateMap['x'] = ALLELSE
		bs.StateMap['.'] = ALLELSE
		TOKEN.FullString = TOKEN.Str
		if TOKEN.HexFlag {
			TOKEN.Isum = FromHex(TOKEN.Str)
		}
		if NEGATV {
			TOKEN.Isum = -TOKEN.Isum
			TOKEN.FullString = "-" + TOKEN.Str
		}
		return TOKEN, false
	}

	//  For OP tokens, must return the opcode as the sum value.  Do this by calling GETOPCODE.
	if TOKEN.State == OP {
		TOKEN.Isum = bs.GETOPCODE(TOKEN)
	}
	return TOKEN, EOL
} // GetToken

//--------------------------------------------------------- GETTKN --------------------------------------

func (bs *BufferState) GETTKN() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = bs.GetToken(true)
	return TOKEN, EOL
} // GETTKN

// ---------------------------------- isdigit -----------------------------------------------
func isdigit(ch rune) bool {
	isdgt := ch >= Dgt0 && ch <= Dgt9
	return isdgt
}

// ---------------------------------- ishexdigit -------------------------------------------------
func ishexdigit(ch rune) bool {

	ishex := isdigit(ch) || ((ch >= 'A') && (ch <= 'F'))
	return ishex

} // ishexdigit

//----------------------------------- fromhex -------------------------------------------------

func FromHex(s string) int {
	result := 0
	var dgtval int
	const OrdinalCapA = 'A'

	for _, dgtchar := range s {
		if isdigit(dgtchar) {
			dgtval = int(dgtchar) - Dgt0
		} else if ishexdigit(dgtchar) {
			dgtval = int(dgtchar) - OrdinalCapA + 10
		} // ignore blanks or any other non digit character.  This includes ignoring the trailing 'H'.
		result = 16*result + dgtval
	}
	return result
} // FromHex

// ---------------------------------------- SetMapDelim -----------------------------------------

func (bs *BufferState) SetMapDelim(char byte) {
	bs.StateMap[char] = DELIM
} // SetMapDelim

//-------------------------------------------- TokenReal ---------------------------------------
// Allows "0x" as hex prefix, as well as "H" as hex suffix.  And idiomatic Go does not have a function begin with Get.

func (bs *BufferState) TokenReal() (TokenType, bool) {
	var token TokenType
	var EOL bool
	var err error

	// I'm hoping to make this routine much less complex, by changing the state of a few characters.
	bs.StateMap['.'] = DGT
	bs.StateMap['_'] = DGT
	wantReal = true

	token, EOL = bs.GETTKN()
	if EOL && token.State != DELIM {
		EOL = false
	}
	if EOL {
		return token, EOL
	}

	if token.State == DGT {
		token.FullString = strings.ReplaceAll(token.FullString, "_", "-")
		token.Rsum, err = strconv.ParseFloat(token.FullString, 64) // FullString field now includes the sign character, if given.
		if err != nil {
			fmt.Printf(" in TokenReal after call to strconv.ParseFloat(%s, 64).  err = %s\n", token.Str, err)
		}
		if token.HexFlag {
			token.Rsum = float64(token.Isum)
		}
	}
	bs.StateMap['_'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
	bs.StateMap['H'] = ALLELSE
	bs.StateMap['E'] = ALLELSE
	bs.StateMap['X'] = ALLELSE
	bs.StateMap['h'] = ALLELSE
	bs.StateMap['e'] = ALLELSE
	bs.StateMap['x'] = ALLELSE
	bs.StateMap['.'] = ALLELSE
	//bs.StateMap['-'] = OP      // make sure the minus sign is back to the type it's supposed to be.
	wantReal = false
	return token, EOL

} // TokenReal

//-------------------------------------------- GETTKNREAL ---------------------------------------
// I am copying the working code from TKNRTNS here.  See the comments in tknrtnsa.adb for reason why.
// Now allows "0x" as hex prefix, as well as "H' as hex suffix.

func (bs *BufferState) GETTKNREAL() (TOKEN TokenType, EOL bool) {
	var CHAR CharType

	TOKEN, EOL = bs.GETTKN()
	if EOL {
		return TOKEN, EOL
	}

	Len := len(TOKEN.Str)
	if (TOKEN.State == ALLELSE) && (Len > 1) && (TOKEN.Str[0] == '.') && isdigit(rune(TOKEN.Str[1])) {
		// Likely have a real number beginning with a decimal point, so fall thru to the digit token
	} else if TOKEN.State != DGT {
		return TOKEN, EOL
	}
	//
	// Now must have a digit token.
	//
	tokenByteSlice := make([]byte, 0, 200) // to build up the TOKEN.Str field
	bs.UNGETTKN()
	TOKEN = TokenType{} // assign nil struct literal to zero all of the fields.
	TOKEN.State = DGT
	bs.PREVPOSN = bs.CURPOSN
	HexFlag := false

ExitLoop:
	for {
		CHAR, EOL = bs.GETCHR()
		CHAR.Ch = CAP(CHAR.Ch)
		if EOL {
			// If TKNSTATE is DELIM, then GETTKN was called when there were
			// no more tokens on line.  Otherwise it means that we have fetched the last token on this line.
			if (TOKEN.State == DELIM) && (len(tokenByteSlice) == 0) {
				break // with EOL still being true.
			} else { // now on last token of line, so don't return with EOL set to true.
				EOL = false
			} // if token state is a delim and token str is empty
		} // IF EOL

		Len = len(tokenByteSlice)
		switch CHAR.State {
		case DELIM: // Ignore leading delims
			if Len > 0 {
				break ExitLoop
			} //goto ExitLoop; }
		case OP:
			if ((CHAR.Ch != '+') && (CHAR.Ch != '-')) || ((Len > 0) && (tokenByteSlice[Len-1] != 'E')) {
				bs.UNGETCHR()
				break ExitLoop // goto ExitLoop;
			}
			tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
		case DGT:
			tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
		case ALLELSE:
			if (CHAR.Ch != '.') && (CHAR.Ch != 'E') && !ishexdigit(rune(CHAR.Ch)) && (CHAR.Ch != 'H') &&
				(CHAR.Ch != 'X') {
				bs.UNGETCHR()
				break ExitLoop // goto ExitLoop;
			} else if CHAR.Ch == 'X' { // have "0x" prefix for a hex number
				HexFlag = true
			} else { // this else clause is so that the 'X' of the "0x" prefix does not get appended.
				tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
			}
		} // Char State
	} // getting characters loop

	TOKEN.DelimCH = CHAR.Ch
	TOKEN.DelimState = CHAR.State
	Len = len(tokenByteSlice)
	TOKEN.Str = string(tokenByteSlice) // An initial assignment that can be changed below.
	if tokenByteSlice[Len-1] == 'H' {
		TOKEN.Str = string(tokenByteSlice[:Len-1]) //  must remove the 'H' from .Str field.
		HexFlag = true
	}
	if HexFlag {
		TOKEN.Isum = FromHex(TOKEN.Str)
		TOKEN.Rsum = float64(TOKEN.Isum)
	} else {
		if tokenByteSlice[0] == '.' {
			TOKEN.Str = "0" + string(tokenByteSlice) // insert leading 0 if number begins with a decimal point.
		}
		TOKEN.Rsum, _ = strconv.ParseFloat(TOKEN.Str, 64) // If err not nil, R becomes 0 by this routine.  I don't have to do it.
		if math.Signbit(TOKEN.Rsum) {
			TOKEN.Isum = int(TOKEN.Rsum - 0.5)
		} else if TOKEN.Rsum == 0.0 {
			TOKEN.Isum = 0
		} else {
			TOKEN.Isum = int(TOKEN.Rsum + 0.5)
		} // if Rsum is negative, zero or positive
	}
	/*
	   I want EOL to only return TRUE when there is no token to process, so I have to make sure that EOL
	   is set to false if there is a token here that was processed.
	*/
	if EOL && TOKEN.State != DELIM {
		EOL = false
	}

	return TOKEN, EOL
} // GETTKNREAL

// --------------------------------------- GetTokenString ---------------------------------------

func (bs *BufferState) GetTokenString(UpperCase bool) (TOKEN TokenType, EOL bool) {
	var Char CharType
	for c := Dgt0; c <= Dgt9; c++ {
		bs.StateMap[byte(c)] = ALLELSE
	}

	// remember that these map assignments could have been altered by SetMapDelim.  In that case I will not change the StateMap for that character
	if bs.StateMap['#'] == OP {
		bs.StateMap['#'] = ALLELSE
	}

	if bs.StateMap['*'] == OP {
		bs.StateMap['*'] = ALLELSE
	}

	if bs.StateMap['+'] == OP {
		bs.StateMap['+'] = ALLELSE
	}

	if bs.StateMap['-'] == OP {
		bs.StateMap['-'] = ALLELSE
	}

	if bs.StateMap['/'] == OP {
		bs.StateMap['/'] = ALLELSE
	}

	if bs.StateMap['<'] == OP {
		bs.StateMap['<'] = ALLELSE
	}

	if bs.StateMap['='] == OP {
		bs.StateMap['='] = ALLELSE /* plussign */
	}

	if bs.StateMap['>'] == OP {
		bs.StateMap['>'] = ALLELSE /* minussign */
	}

	TOKEN, EOL = bs.GetToken(UpperCase)
	if EOL || (TOKEN.State == DELIM) || ((TOKEN.State == ALLELSE) && (TOKEN.DelimState == DELIM)) {
		return // TOKEN,EOL;
	}

	/*
	   Now must do special function of this proc.
	   Continue building the string as left off from GetToken call.
	   As of 6/95 this should always return a tknstate of allelse, so return.
	*/

	tokenByteSlice := make([]byte, 0, 200)
	copy(tokenByteSlice, TOKEN.Str)
	for {
		Char, EOL = bs.GETCHR() // the CAP function is not here anymore.
		if EOL || ((Char.State == DELIM) && (len(TOKEN.Str) > 0)) {
			break // Ignore leading delims
		}
		tokenByteSlice = append(tokenByteSlice, Char.Ch)
		TOKEN.Isum += int(Char.Ch)
	} // getting chars
	TOKEN.Str = string(tokenByteSlice)

	if UpperCase {
		TOKEN.Str = strings.ToUpper(TOKEN.Str)
	}
	TOKEN.Rsum = float64(TOKEN.Isum)
	TOKEN.DelimCH = Char.Ch
	TOKEN.DelimState = Char.State
	if EOL && TOKEN.State != DELIM {
		EOL = false
	}
	return TOKEN, EOL
} // GetTokenString

// --------------------------------------- GETTKNSTR ---------------------------------------

func (bs *BufferState) GETTKNSTR() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = bs.GetTokenString(true)
	return TOKEN, EOL
}

// ---------------------------------------- GetTokenEOL -------------------------------------------

func (bs *BufferState) GetTokenEOL(UpperCase bool) (TOKEN TokenType, EOL bool) {
	// GET ToKeN to EndOfLine.
	// This will build a token that consists of every character left on the line.
	// That is, it only stops at the end of line.
	// The TRIM procedure is used to set the COUNT and LENGTH fields.  This is
	// the only TOKENIZE procedure that uses it.

	var Char CharType
	bs.PREVPOSN = bs.CURPOSN // So this tkn can be ungotten as well
	tokenByteSlice := make([]byte, 0, 200)
	TOKEN = TokenType{}
	for {
		Char, EOL = bs.GETCHR() // the Cap function is not here anymore.
		if EOL {
			break
		} //  No-go.  Need to change this to idiomatic Go, but after debugging the main GetTkn and UnGetTkn rtns.
		tokenByteSlice = append(tokenByteSlice, Char.Ch)
		//                                                                        TOKEN.Str += string(Char.Ch);
		TOKEN.Isum += int(Char.Ch)
		TOKEN.State = ALLELSE
	}
	TOKEN.Str = string(tokenByteSlice)

	if UpperCase {
		TOKEN.Str = strings.ToUpper(TOKEN.Str)
	}
	TOKEN.Rsum = float64(TOKEN.Isum)
	TOKEN.DelimCH = NullChar
	TOKEN.DelimState = DELIM
	if EOL && len(TOKEN.Str) > 0 {
		EOL = false
	}
	return TOKEN, EOL
} // GetTokenEOL

// ----------------------------------------- GETTKNEOL ------------------------------------------

func (bs *BufferState) GETTKNEOL() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = bs.GetTokenEOL(true)
	return TOKEN, EOL
} // GETTKNEOL

//  UNGETTKN is an internal function

func (bs *BufferState) UNGETTKN() {
	/*
	   * UNGET TOKEN ROUTINE.
	   This routine will unget the last token fetched.  It does this by restoring
	   the previous value of POSN, held in PREVPOSN.  Only the last token fetched
	   can be ungotten, so PREVPOSN is reset after use.  If PREVPOSN contains this
	   as its value, then the unget operation will fail.
	*/
	if (bs.CURPOSN <= bs.PREVPOSN) || (bs.PREVPOSN < 0) {
		log.SetFlags(log.Llongfile)
		log.Print("CurPosn out_of_range in UnGetTkn")
		os.Exit(1)
	} // End error trap

	bs.CURPOSN = bs.PREVPOSN
	bs.PREVPOSN = 0
}

//                                        GetTokenSlice, now TokenSlice

func TokenSlice(str string) []TokenType {
	if str == "" {
		return nil
	}
	bs := New(str)                        // bs is a buffer state
	tknslice := make([]TokenType, 0, 100) // arbitrary limit, ie, a magic number as per Rob Pike.

	for {
		tkn, eol := bs.GetToken(false)
		if eol {
			break
		}
		tknslice = append(tknslice, tkn)
	}
	return tknslice
}

//                                 Old RealTokenSlice

func RealTokenSlice(str string) []TokenType {
	if str == "" {
		return nil
	}
	bufstate := New(str)
	realtknslice := make([]TokenType, 0, 100)

	for {
		tknreal, eol := bufstate.GETTKNREAL()
		if eol {
			break
		}
		realtknslice = append(realtknslice, tknreal)
	}
	return realtknslice
}

// -------------------------- TokenRealSlice ------------------------------

func TokenRealSlice(str string) []TokenType { // This uses the new TokenReal instead of the old GETTKNREAL.
	if str == "" {
		return nil
	}
	bufState := New(str)
	realTknSlice := make([]TokenType, 0, 10)

	for {
		tknreal, eol := bufState.TokenReal()
		if eol {
			break
		}
		realTknSlice = append(realTknSlice, tknreal)
	}
	return realTknSlice
}

// end tknptr

/*
  A way to output program file and line numbers in an error message.  Must be a closure, I think.  But
  this is cumbersome.  I'll leave the code here in case I figure out a way to make it less cumbersome.
  where := func() {
      _, file, line, _ := runtime.Caller(1)
      log.Fatalf(" In UNGETCHR and CurPosn is < 0.  %s:%d\n", file, line)
  }
  where();
*/
