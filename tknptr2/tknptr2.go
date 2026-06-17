package tknptr2 // Package tknptr2 from tknptrutf8 from tknptr.

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

/*
 Copyright (C) 1987-2026  Robert Solomon MD.  All rights reserved.
 These routines collectively implement a very good facility to fetch, manipulate, and interpret tokens.

REVISION HISTORY
----------------
1987      -- Modula-2 section
28 MAY 87 -- Added UNGETTKN capability and no longer exported GETCHR and UNGETCHR.
29 AUG 87 -- Restored exportation of GETCHR and UNGETCHR.
 3 Mar 88 -- Added the ASCZERO declaration and removed the function call from the DGT conversion loop.
31 Mar 88 -- Converted to M2 V3.03.
 1 Sept 88 -- 1.  Allowed quoted string to force ALLELSE state.
              2.  Changed the method UNGETTKN uses to unget the token.
              3.  Added the MULTSIGN and DIVSIGN OP states.
              4.  Ran M2CHECK and deleted all unreferenced var's.
              5.  Moved the NEGATV check for contingently making SUM < 0 out
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
----------------------------------------------------------------------------------------------------
 9 Oct 13 -- Converted to gm2.
11 Oct 13 -- Fixed a bug in GETTKNREAL in which number like 1e-1 lost the e.
12 Oct 13 -- Removed an errant RETURN from GETTKNSTR.
----------------------------------------------------------------------------------------------------
 3 Feb 14 -- Converted to Ada.  I modernized the data types to be a record type.
28 Jun 14 -- Backported enhancement to GetOpCode that includes ^, ** and %.
----------------------------------------------------------------------------------------------------
19 Nov 14 -- Converted to C++.
 7 Dec 14 -- Removed comma as a delim, making it AllElse so it works as intended for HPCALCC
28 Dec 14 -- Turns out that CentOS C++ does not support -std=c++11, so I have to remove string.front and string.back member functions.
18 Jan 15 -- Found bug in which single digits followed by add or subtract are not processed correctly by GETTKNREAL.
----------------------------------------------------------------------------------------------------
 6 Aug 16 -- Started conversion to Go, while on board boat to Bermuda.
19 Aug 16 -- Finished conversion to Go
21 Sep 16 -- Now that this code is for case-sensitive filesystem like linux, returning an all-caps token is a bad idea.
               So I added FetchToken, which takes a param of true for cap and false for preserving case.
 9 Oct 16 -- Will allow "0x" as prefix for hex, as well as "H" suffix.  An 'x' anywhere in the number will
                be a hex number.  I will not force it to be the 2nd character.
25 Nov 16 -- The TKNMAXSIZ was too small for sha512, so I increased it.
 3 Dec 16 -- Decided to change how the UpperCase flag is handled in GetToken.
10 Aug 17 -- Making this use pointer receivers, if I can.
13 Oct 17 -- Made tab char a delim.  Needed for comparehashes.
18 Oct 17 -- Changed init process, so all control codes are delims, just as in the current tokenize.
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
19 Jul 23 -- Added TokenRealSlice, which uses the new TokenReal(), instead of the old GETTKNREAL().  And I decided to not allow '-' for negative exponents.  I very rarely use E notation,
               so I want to preserve the use of '-' as an operator.  Underscore, '_', is replaced w/ '-' just before call to strconv.ParseFloat().
23 Jul 23 -- The code is now working as designed, as tested w/ tknptr_test.go and testtokenptr.go.
               The old GETTKNREAL works by using GetToken() to determine the state of the next token.  If it's not a number, that result is returned.  If it's a number, ungettoken is called
               and a real number token is parsed from scratch.  That repeats much of the finite state logic.
               The new TokenReal() works by changing the StateMap for some characters and then using a slightly modified GetToken to get a real number (float64).  This makes the code
               much simpler; the only use of the finite state automaton is in GetToken().
               And I also added 3 fields to TokenType (FullString, RealFlag and HexFlag), and added a signaling flag, wantReal that TokenReal() uses to signal into GetToken().
24 Jul 23 -- Spoke too soon.  Hex input isn't working correctly.  Gotta fix that now.  And I removed 'h' to indicate hex.  Now only 0x will work, as used in C-ish.
24 May 24 -- Added comments that will be detected by go doc
10 Sep 25 -- Added a stringer method for TokenType
------------------------------------------------------------------------------------------------------------------------------------------------------
11 Jun 26 -- Now called tknptrutf8, so it will use UTF-8.  I'll stop assuming that a character is a byte.
13 Jun 26 -- Yesterday I used utf8.RuneLen() to increment the pointer to the next character.  That's a mistake as I'm now using runes instead of bytes.  I'll revert back to
				incrementing or decrementing the pointer by 1.
------------------------------------------------------------------------------------------------------------------------------------------------------
14 Jun 26 -- Now called tknptr2, so it will use UTF-8.  I'll stop assuming that a character is a byte.  I'm going to play with getting string builder and string reader to work.
			Because I need ungetchar and ungettoken, this may be more work than I need.  So I won't use a string reader for the source runes, but I will use a string builder for the token.
			It passes the tests in tknptr2_test.go, and also in testtokenptr2.
15 Jun 26 -- I'm removing HoldLineBS as it's not needed.  That worked.  Now I'm removing HOLDCURPOSN as it's not needed either.  And STOTKNPOSN and RCLTKNPOSN are not used.
				These all are from the Modula-2 days, and I never examined them.  Time to remove them.
17 Jun 26 -- I'm going to play with getting string reader to work.  I'll need 2 string readers in the bufferState, so I can unget a token.  This is the hard part, ungetting a token.
				I got it to mostly work.  In that go test fails some cases, but testtokenptr2 passes.  I'll stop for now.  I'm going to use tknptr or tknptrutf8 instead.
*/

const LastAltered = "17 June 2026"

const (
	DELIM = iota // so DELIM = 0, and so on.  And the zero val needs to be DELIM.
	OP
	DGT
	ALLELSE
)

// TokenType fields are Str; FullStr which includes the minus sign; and others
type TokenType struct {
	Str        string
	FullString string // includes minus sign character, if present.
	State      int
	DelimCH    rune
	DelimState int
	Isum       int
	Rsum       float64
	RealFlag   bool // flag so integer processing stops when it sees a dot, E or e.
	HexFlag    bool // only way I know of to signal that the input string is a hex format.
}

// CharType fields are Ch rune and State int.
type CharType struct {
	Ch    rune
	State int
}

type BufferState struct {
	CURPOSN, PREVPOSN int
	// lineRuneSlice     []rune
	strReader1, strReader2 *strings.Reader
	StateMap               map[rune]int // as of 9/28/20, StateMap is part of this structure.
}

var FSAnameType = [...]string{"DELIM", "OP", "DGT", "ALLELSE"}

func (t TokenType) String() string { // satisfies the stringer interface for TokenType
	var s string
	if t.Rsum == 0 {
		s = fmt.Sprintf("Str: %s, fullStr: %s, State: %s, DelimCh: 0x%02X, DelimState: %s, Isum: %d, Rsum: %g, RealFlag: %t, HexFlag: %t",
			t.Str, t.FullString, FSAnameType[t.State], t.DelimCH, FSAnameType[t.DelimState], t.Isum, t.Rsum, t.RealFlag, t.HexFlag)
	} else if math.Abs(t.Rsum) < 1e8 {
		s = fmt.Sprintf("Str: %s, fullStr: %s, State: %s, DelimCh: %0#2X, DelimState: %s, Isum: %d, Rsum: %.2f, RealFlag: %t, HexFlag: %t",
			t.Str, t.FullString, FSAnameType[t.State], t.DelimCH, FSAnameType[t.DelimState], t.Isum, t.Rsum, t.RealFlag, t.HexFlag)
	} else {
		s = fmt.Sprintf("Str: %s, fullStr: %s, State: %s, DelimCh: 0x%02X, DelimState: %s, Isum: %d, Rsum: %.4g, RealFlag: %t, HexFlag: %t",
			t.Str, t.FullString, FSAnameType[t.State], t.DelimCH, FSAnameType[t.DelimState], t.Isum, t.Rsum, t.RealFlag, t.HexFlag)
	}
	return s
}

// ---------------------------------------------------------------------

const TKNMAXSIZ = 256
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

// Cap -- will convert a rune to its upper case value.  It takes a rune and returns a rune.  Not used.
//func Cap(c rune) rune {
//	r, _, _, _ := strconv.UnquoteChar(strings.ToUpper(string(c)), 0)
//	return r
//} // Cap

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

// CAP -- will convert the param to an upper case letter.  It takes a byte and returns a byte.
func CAP(c rune) rune {
	if (c >= 'a') && (c <= 'z') {
		c = c - 32
	}
	return c
}

// ------------------------------------ InitStateMap ------------------------------------------------
// Making sure that the StateMap is at its default values, since a call to GetTokenStr changes some values.

func InitStateMap(bs *BufferState) {
	bs.StateMap = make(map[rune]int, 128)
	//	StateMap[NullChar] = DELIM
	for i := 0; i < 33; i++ { // note that this includes \t, HT, tab character.
		bs.StateMap[rune(i)] = DELIM
	}
	for i := 33; i < 128; i++ {
		bs.StateMap[rune(i)] = ALLELSE // including comma
	}
	for c := Dgt0; c <= Dgt9; c++ {
		bs.StateMap[c] = DGT // character 'c' is now a rune, so it doesn't need to be converted to a rune
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

// ----------------------------------------- New ----------------------------------------

// New -- input is a string, and it returns a *BufferState.  This is the idiomatic function call to start tokenizing.  The others have been removed.
func New(Str string) *BufferState { // constructor, initializer using idiomatic Go as taught by Bill Kennedy and others.
	var bufState BufferState

	//bs := new(BufferState) // idiomatic Go could write this as &BufferState{}  And I stopped using bs as byteslice because it looks ugly, like bullshit.
	InitStateMap(&bufState) // possible that GetTknStr or GetTknEOL changed the StateMap, so will call init.
	// bufState.CURPOSN, bufState.PREVPOSN, bufState.HOLDCURPOSN = 0, 0, 0  Don't need HOLDCURPOSN.
	bufState.CURPOSN, bufState.PREVPOSN = 0, 0 // not needed in Go, carried over from Modula-2 then Ada then C++
	//       bufState.lineRuneSlice = []rune(Str)
	bufState.strReader1 = strings.NewReader(Str)
	bufState.strReader2 = strings.NewReader(Str)
	return &bufState // makes clear that the return value is a pointer to a BufferState, and uses pointer semantics.
} //

// ---------------------------- GetChar -------------------------------------------

// GetChar -- Gets the next character in the buffer state, returning a CharType and a bool.  The bool is intended for the EOL condition.
func (bufState *BufferState) GetChar() (CharType, bool) {
	var c CharType
	var EOL bool
	var err error

	// c.Ch = bs.lineRuneSlice[bs.CURPOSN] //  no longer use the Cap function here.
	c.Ch, _, err = bufState.strReader1.ReadRune()
	if err != nil {
		return c, true
	}
	c.State = bufState.StateMap[c.Ch] // state assignment, here using map access.
	bufState.CURPOSN++
	return c, EOL
} // PeekCHR

// NextChr -- only increments the Current position index.  Not used now that I'm using string reader.
//func (bs *BufferState) NextChr() {
//	bs.CURPOSN++
//} // NextChr
// GETCHR -- Does both a PeekChr and a NextChr, returning a CharType and a bool which is true when EOL condition is met.
//func (bs *BufferState) GETCHR() (CharType, bool) {
//	C, EOL := bs.PeekChr()
//	bs.NextChr()
//	return C, EOL
//} // GETCHR

// --------------------------------- UnGetChar --------------------------------

// UnGetChar -- Does what its name says.  Primarily an internal function that decrements the current position index.
func (bufState *BufferState) UnGetChar() {
	err := bufState.strReader1.UnreadRune()
	if err != nil {
		log.SetFlags(log.Llongfile)
		log.Print("Error in UnGetChar: ", err, ", CURPOSN=", bufState.CURPOSN, ", PrevPosn=", bufState.PREVPOSN)
		//os.Exit(1)
		fmt.Println()
	}
	bufState.CURPOSN--
} // UnGetChar

// ------------------------------------- GetOpCode ---------------------------------------------

// GETOPCODE -- Does what its name says.  Primarily an internal function.
func (bufState *BufferState) GETOPCODE(Token TokenType) int {

	//-- GET OPCODE.
	//-- This routine receives a token of FSATYP op (meaning it is an operator)
	//-- and analyzes it to determine an opcode, which is a nUMBER from 1..22.
	//-- This is done after the necessary validity check of the input token.
	//-- The opcode assignments for the op tokens are:
	//--  < is 1                  <= is 2
	//--  > is 3                  >= is 4
	//--  = is 5   == is 5        <> is 6    # is 7
	//--  + is 8                  += is 9
	//--  - is 10                 -= is 11
	//--  * is 12                 *= is 13
	//--  / is 14                 /= is 15
	//--  ^ is 16                 ^= is 17
	//-- ** is 18                **= is too long to be allowed
	//-- >< is 20
	//--  % is 22

	var CH1, CH2 byte
	OpCode := 0

	if len(Token.Str) < 1 {
		log.SetFlags(log.Llongfile)
		log.Print(" Token is empty, from GetOpCode.")
		return OpCode
	}

	// keep token length to 2 chars max.
	for len(Token.Str) > 2 {
		bufState.UnGetChar()
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
		bufState.UnGetChar() // unget the 2nd part of the invalid pair.
	} // Length of Token = 1
	return OpCode
} // GETOPCODE

//       ---------------------------=== GetToken ===--------------------------------------

// GetToken (UpperCase bool) -- returns a TokenType which may be uppercase; EOL will be true when EOL condition is met.  Does not process scientific notation.
func (bufState *BufferState) GetToken(UpperCase bool) (TOKEN TokenType, EOL bool) {

	var CHAR CharType
	var QUOCHR rune // Holds the active quote char

	bufState.PREVPOSN = bufState.CURPOSN
	// bufState.strReader2 = bufState.strReader1 // hope this just doesn't copy a pointer.  Nope, it does a shallow copy.
	bufState.strReader2 = bufState.strReader1
	_, err := bufState.strReader2.Seek(int64(bufState.CURPOSN), io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}
	var NEGATV, QUOFLG bool
	TOKEN = TokenType{} // This will zero out all the fields by using a nil struct literal.  It's the default; I put it here so I remember.
	//tokenRuneSlice := make([]rune, 0, 200) // to build up the TOKEN.Str field
	var buildingToken strings.Builder
	buildingToken.Grow(200)

ExitForLoop:
	for {
		CHAR, EOL = bufState.GetChar()
		if EOL {
			// If TKNSTATE is DELIM, then gettkn was called when there were no more tokens on the input line.
			// Otherwise, it means that we have fetched the last TOKEN on this line.
			if (TOKEN.State == DELIM) && (buildingToken.Len() == 0) {
				break // with EOL still being true.
			} else { // now on last token of line, so don't return with EOL set to true.
				EOL = false
			}
		}
		if QUOFLG && (CHAR.Ch != NullChar) {
			CHAR.State = ALLELSE
		}

		switch TOKEN.State {
		case DELIM: // token.state
			switch CHAR.State {
			case DELIM: // NULL char is a special delimiter because it will
				// immediately cause a return even if there is no token yet, i.e., the token is only delimiters.  This is because of
				// the NULL char is the string terminator for general strings and especially for environment strings, for which this
				// TOKENIZE module was originally written.
				if CHAR.Ch == NullChar {
					break ExitForLoop
				} // goto ExitLoop }
			case OP: // Delim -> OP means this is the 1st char of the operator token.
				TOKEN.State = OP
				//tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
				buildingToken.WriteRune(CHAR.Ch)
				if buildingToken.Len() > TKNMAXSIZ {
					log.SetFlags(log.Llongfile)
					log.Println(" token too long in GetToken.")
					os.Exit(1)
				}
			case DGT: // Delim -> DGT means this is the 1st dgt for the entered number.
				//tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
				buildingToken.WriteRune(CHAR.Ch)
				TOKEN.State = DGT
				if wantReal {
					bufState.StateMap['E'] = DGT
					bufState.StateMap['e'] = DGT
				}
				if !unicode.IsDigit(CHAR.Ch) {
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
					//tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
					buildingToken.WriteRune(CHAR.Ch)
					if buildingToken.Len() > TKNMAXSIZ {
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
				bufState.UnGetChar() // To allow correct processing of op pair that is not a valid op, like +- or =>
				break ExitForLoop    //goto ExitLoop;
			case OP: // OP -> OP means another operator character found.
				if buildingToken.Len() > OpMaxSize {
					bufState.UnGetChar()
					break ExitForLoop //goto ExitLoop;
				}
				//tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
				buildingToken.WriteRune(CHAR.Ch)
			case DGT: // OP -> DGT means it may be a sign character for a number token.  If not, have 1 char operator
				upperbound := buildingToken.Len() - 1
				LastChar := buildingToken.String()[upperbound]
				if (LastChar == '+') || (LastChar == '-') {
					if buildingToken.Len() == 1 {
						if buildingToken.String()[0] == '-' {
							NEGATV = true
						}
						TOKEN.State = DGT
						// OVERWRITE ARITHMETIC SIGN CHARACTER
						//tokenRuneSlice[0] = CHAR.Ch
						buildingToken.Reset()
						buildingToken.WriteRune(CHAR.Ch)
						if wantReal {
							bufState.StateMap['E'] = DGT
							bufState.StateMap['e'] = DGT
						}
						if !unicode.IsDigit(CHAR.Ch) {
							TOKEN.RealFlag = true
							continue
						}
						TOKEN.Isum = int(CHAR.Ch) - Dgt0
					} else { // TOKEN length is not 1 so must first return valid OP
						bufState.UnGetChar()     // UNGET THIS DIGIT CHAR
						bufState.UnGetChar()     // THEN UNGET THE ARITH SIGN CHAR
						CHAR.Ch = rune(LastChar) // SO DELIMCH CORRECTLY RETURNS THE ARITH SIGN CHAR
						//tokenRuneSlice = tokenRuneSlice[:upperbound] // recall that upperbound is excluded in this syntax, so this removes the last char of the token
						// to remove the last character, I need a few lines of code here.  When I was using rune slice syntax, this just needed one line of code
						tempStr := buildingToken.String()[:upperbound]
						buildingToken.Reset()
						buildingToken.WriteString(tempStr)

						break ExitForLoop //goto ExitLoop;
					} // if length of the token = 1
				} else { // IF have a sign character as the last char
					bufState.UnGetChar()
					break ExitForLoop //goto ExitLoop;
				} // If have a sign character as the last char
			case ALLELSE: // OP -> AllElse
				bufState.UnGetChar()
				break ExitForLoop //goto ExitLoop;
			} // Char.State
		case DGT: // tokenstate
			switch CHAR.State {
			case DELIM:
				break ExitForLoop
			case OP: // DGT -> OP
				bufState.UnGetChar()
				bufState.StateMap['_'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
				break ExitForLoop                //goto ExitLoop;
			case DGT: // DGT -> DGT so we have another digit.
				//tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
				buildingToken.WriteRune(CHAR.Ch)
				if CAP(CHAR.Ch) == 'E' {
					bufState.StateMap['_'] = DGT // make the underscore to be of DGT type so it will be allowed in the number
					//bs.StateMap['-'] = DGT // make the minus sign to be of DGT type so it will be allowed in the number.  Nope, changed my mind.  I almost never use E notation.
				}
				if TOKEN.RealFlag { // Isum only will contain the int part of a float.
					continue
				}
				if wantReal {
					bufState.StateMap['E'] = DGT
					bufState.StateMap['e'] = DGT
				}
				if !unicode.IsDigit(rune(CHAR.Ch)) {
					TOKEN.RealFlag = true
					continue
				}
				TOKEN.Isum = 10*TOKEN.Isum + int(CHAR.Ch) - Dgt0 // this total will not be correct when floating point chars, like dot and 'E' or 'e', are input.
			case ALLELSE: // DGT -> AllElse
				if rune(CHAR.Ch) == 'x' || rune(CHAR.Ch) == 'X' {
					TOKEN.HexFlag = true
					bufState.StateMap['a'] = DGT
					bufState.StateMap['b'] = DGT
					bufState.StateMap['c'] = DGT
					bufState.StateMap['d'] = DGT
					bufState.StateMap['e'] = DGT
					bufState.StateMap['f'] = DGT
					bufState.StateMap['A'] = DGT
					bufState.StateMap['B'] = DGT
					bufState.StateMap['C'] = DGT
					bufState.StateMap['D'] = DGT
					bufState.StateMap['E'] = DGT
					bufState.StateMap['F'] = DGT
					continue
				}

				bufState.UnGetChar()
				break ExitForLoop //goto ExitLoop;
			} // Char.State
		case ALLELSE: // tokenstate
			switch CHAR.State {
			case DELIM:
				//  Always exit if get a NULL char as a delim.  A quoted string can only get here if CH is NULL.
				break ExitForLoop //goto ExitLoop;
			case OP:
				bufState.UnGetChar()
				break ExitForLoop //goto ExitLoop;
			case DGT: // AllElse -> DGT means have alphanumeric token.
				if buildingToken.Len() > TKNMAXSIZ {
					log.SetFlags(log.Llongfile)
					log.Println(" token too long in GetTkn AllELSE to Digit branch.")
					os.Exit(1)
				} // if token too long
				//tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
				buildingToken.WriteRune(CHAR.Ch)
				TOKEN.Isum += int(CHAR.Ch)
			case ALLELSE: // AllElse -> AllELSE
				if rune(CHAR.Ch) == QUOCHR {
					QUOFLG = false
					CHAR.State = DELIM // So that DELIMSTATE will = delim
					break ExitForLoop

				} else {
					if buildingToken.Len() > TKNMAXSIZ {
						log.SetFlags(log.Llongfile)
						log.Println(" token too long in GetTkn AllElse -> AllElse branch.")
					} // if token too long
					//tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
					buildingToken.WriteRune(CHAR.Ch)
					TOKEN.Isum += int(CHAR.Ch)
				} // if char is a quote char
			} // Char.State
		} // Token.State
	} //LOOP to process characters

	if UpperCase {
		TOKEN.Str = strings.ToUpper(buildingToken.String())
	} else {
		TOKEN.Str = buildingToken.String()
	}
	TOKEN.DelimCH = CHAR.Ch
	TOKEN.DelimState = CHAR.State
	TOKEN.FullString = TOKEN.Str
	if TOKEN.RealFlag {
		TOKEN.Str = strings.ReplaceAll(TOKEN.Str, "_", "-") // Note that '_' could mean '-' for exponents.
	}
	if TOKEN.State == DGT {
		//bs.StateMap['-'] = OP      // make sure the minus sign is back to the type it's supposed to be.
		bufState.StateMap['_'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
		bufState.StateMap['E'] = ALLELSE
		bufState.StateMap['e'] = ALLELSE
		bufState.StateMap['.'] = ALLELSE
		TOKEN.FullString = TOKEN.Str
		if TOKEN.HexFlag {
			TOKEN.Isum = FromHex(TOKEN.Str)
			bufState.StateMap['a'] = ALLELSE
			bufState.StateMap['b'] = ALLELSE
			bufState.StateMap['c'] = ALLELSE
			bufState.StateMap['d'] = ALLELSE
			bufState.StateMap['e'] = ALLELSE
			bufState.StateMap['f'] = ALLELSE
			bufState.StateMap['A'] = ALLELSE
			bufState.StateMap['B'] = ALLELSE
			bufState.StateMap['C'] = ALLELSE
			bufState.StateMap['D'] = ALLELSE
			bufState.StateMap['E'] = ALLELSE
			bufState.StateMap['F'] = ALLELSE
		}
		if NEGATV {
			TOKEN.Isum = -TOKEN.Isum
			TOKEN.FullString = "-" + TOKEN.Str
		}
		return TOKEN, false
	}

	//  For OP tokens, must return the opcode as the sum value.  Do this by calling GETOPCODE.
	if TOKEN.State == OP {
		TOKEN.Isum = bufState.GETOPCODE(TOKEN)
	}
	return TOKEN, EOL
} // GetToken

//--------------------------------------------------------- GETTKN --------------------------------------

// GETTKN -- returns an upper-cased token and EOL condition.
func (bufState *BufferState) GETTKN() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = bufState.GetToken(true)
	return TOKEN, EOL
} // GETTKN

// ---------------------------------- isdigit -----------------------------------------------

// isdigit -- input a rune and return a bool.
func isdigit(ch rune) bool {
	isdgt := ch >= Dgt0 && ch <= Dgt9
	return isdgt
}

// ---------------------------------- ishexdigit -------------------------------------------------
// ishexdigit -- input a rune and return a bool.
func ishexdigit(ch rune) bool {

	ishex := isdigit(ch) || ((ch >= 'A') && (ch <= 'F'))
	return ishex

} // ishexdigit

//----------------------------------- FromHex -------------------------------------------------

// FromHex -- input a string and returns int.
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

// SetMapDelim -- input a byte that will be included in the characters that are used as delimiters.
func (bufState *BufferState) SetMapDelim(char rune) {
	bufState.StateMap[char] = DELIM
} // SetMapDelim

//-------------------------------------------- TokenReal ---------------------------------------

// TokenReal allows "0x" as the hex prefix, and it no longer allows "H" as a hex suffix.  And idiomatic Go does not have a function begin with Get.
// TokenReal -- returns a TokenType and EOL state.  Does process scientific notation by changing the state of certain characters to simplify the code.
func (bufState *BufferState) TokenReal() (TokenType, bool) {
	var token TokenType
	var EOL bool
	var err error

	// I'm hoping to make this routine much less complex, by changing the state of a few characters.
	bufState.StateMap['_'] = DGT
	bufState.StateMap['.'] = DGT
	wantReal = true

	token, EOL = bufState.GETTKN()
	if EOL && token.State != DELIM {
		EOL = false
	}
	if EOL {
		return token, EOL
	}

	if token.State == DGT {
		token.FullString = strings.ReplaceAll(token.FullString, "_", "-")
		if token.HexFlag {
			token.Rsum = float64(token.Isum)
		} else {
			token.Rsum, err = strconv.ParseFloat(token.FullString, 64) // FullString field now includes the sign character, if given.
		}
		if err != nil {
			fmt.Printf(" in TokenReal after call to strconv.ParseFloat(%s, 64).  err = %s\n", token.Str, err)
		}
	}
	bufState.StateMap['_'] = ALLELSE // make sure the underscore is back to the type it's supposed to be.
	bufState.StateMap['.'] = ALLELSE
	bufState.StateMap['E'] = ALLELSE
	bufState.StateMap['e'] = ALLELSE
	wantReal = false
	return token, EOL

} // TokenReal

//-------------------------------------------- GETTKNREAL ---------------------------------------
// I am copying the working code from TKNRTNS here.  See the comments in tknrtnsa.adb for reason why.
// Allows "0x" as hex prefix; no longer allows "H" as hex suffix.

// GETTKNREAL -- Returns a TokenType and EOL indicater.  This is the rtn to do this.
func (bufState *BufferState) GETTKNREAL() (TOKEN TokenType, EOL bool) {
	var CHAR CharType

	TOKEN, EOL = bufState.GETTKN()
	if EOL {
		return TOKEN, EOL
	}

	Len := len(TOKEN.Str)
	if (TOKEN.State == ALLELSE) && (Len > 1) && (TOKEN.Str[0] == '.') && isdigit(rune(TOKEN.Str[1])) {
		// Likely have a real number beginning with a decimal point, so fall through to the digit token
	} else if TOKEN.State != DGT {
		return TOKEN, EOL
	}
	//
	// Now must have a digit token.
	//
	tokenRuneSlice := make([]rune, 0, 200) // to build up the TOKEN.Str field
	bufState.UNGETTKN()
	TOKEN = TokenType{} // assign nil struct literal to zero all the fields.
	TOKEN.State = DGT
	bufState.PREVPOSN = bufState.CURPOSN
	HexFlag := false

ExitLoop:
	for {
		CHAR, EOL = bufState.GetChar()
		CHAR.Ch = CAP(CHAR.Ch)
		if EOL {
			// If TKNSTATE is DELIM, then GETTKN was called when there were
			// no more tokens on line.  Otherwise, it means that we have fetched the last token on this line.
			if (TOKEN.State == DELIM) && (len(tokenRuneSlice) == 0) {
				break // with EOL still being true.
			} else { // now on last token of line, so don't return with EOL set to true.
				EOL = false
			} // if token state is a delim and token str is empty
		} // IF EOL

		Len = len(tokenRuneSlice)
		switch CHAR.State {
		case DELIM: // Ignore leading delims
			if Len > 0 {
				break ExitLoop
			} //goto ExitLoop; }
		case OP:
			if ((CHAR.Ch != '+') && (CHAR.Ch != '-')) || ((Len > 0) && (tokenRuneSlice[Len-1] != 'E')) {
				bufState.UnGetChar()
				break ExitLoop // goto ExitLoop;
			}
			tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
		case DGT:
			tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
		case ALLELSE:
			if (CHAR.Ch != '.') && (CHAR.Ch != 'E') && !ishexdigit(rune(CHAR.Ch)) && (CHAR.Ch != 'H') &&
				(CHAR.Ch != 'X') {
				bufState.UnGetChar()
				break ExitLoop // goto ExitLoop;
			} else if CHAR.Ch == 'X' { // have "0x" prefix for a hex number
				HexFlag = true
			} else { // this else clause is so that the 'X' of the "0x" prefix does not get appended.
				tokenRuneSlice = append(tokenRuneSlice, CHAR.Ch)
			}
		} // Char State
	} // getting characters loop

	TOKEN.DelimCH = CHAR.Ch
	TOKEN.DelimState = CHAR.State
	Len = len(tokenRuneSlice)
	TOKEN.Str = string(tokenRuneSlice) // An initial assignment that can be changed below.
	if tokenRuneSlice[Len-1] == 'H' {
		TOKEN.Str = string(tokenRuneSlice[:Len-1]) //  must remove the 'H' from .Str field.
		HexFlag = true
	}
	if HexFlag {
		TOKEN.Isum = FromHex(TOKEN.Str)
		TOKEN.Rsum = float64(TOKEN.Isum)
	} else {
		if tokenRuneSlice[0] == '.' {
			TOKEN.Str = "0" + string(tokenRuneSlice) // insert leading 0 if number begins with a decimal point.
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

// GetTokenString (uppercase bool) -- returns a possibly all upper case TokenType and the EOL indicator.
func (bufState *BufferState) GetTokenString(UpperCase bool) (TOKEN TokenType, EOL bool) {
	var Char CharType
	for c := Dgt0; c <= Dgt9; c++ {
		bufState.StateMap[c] = ALLELSE
	}

	// remember that these map assignments could have been altered by SetMapDelim.  In that case I will not change the StateMap for that character
	if bufState.StateMap['#'] == OP {
		bufState.StateMap['#'] = ALLELSE
	}

	if bufState.StateMap['*'] == OP {
		bufState.StateMap['*'] = ALLELSE
	}

	if bufState.StateMap['+'] == OP {
		bufState.StateMap['+'] = ALLELSE
	}

	if bufState.StateMap['-'] == OP {
		bufState.StateMap['-'] = ALLELSE
	}

	if bufState.StateMap['/'] == OP {
		bufState.StateMap['/'] = ALLELSE
	}

	if bufState.StateMap['<'] == OP {
		bufState.StateMap['<'] = ALLELSE
	}

	if bufState.StateMap['='] == OP {
		bufState.StateMap['='] = ALLELSE /* plussign */
	}

	if bufState.StateMap['>'] == OP {
		bufState.StateMap['>'] = ALLELSE /* minussign */
	}

	TOKEN, EOL = bufState.GetToken(UpperCase)
	if EOL || (TOKEN.State == DELIM) || ((TOKEN.State == ALLELSE) && (TOKEN.DelimState == DELIM)) {
		return // TOKEN,EOL;
	}

	// Now must do special function of this proc.
	// Continue building the string as left off from GetToken call.
	// As of 6/95 this should always return a tknstate of allelse, so return.

	tokenRuneSlice := make([]rune, 0, 200)
	copy(tokenRuneSlice, []rune(TOKEN.Str))

	for {
		Char, EOL = bufState.GetChar() // the CAP function is not here anymore.
		if EOL || ((Char.State == DELIM) && (len(TOKEN.Str) > 0)) {
			break // Ignore leading delims
		}
		tokenRuneSlice = append(tokenRuneSlice, Char.Ch)
		TOKEN.Isum += int(Char.Ch)
	} // getting chars
	TOKEN.Str = string(tokenRuneSlice)

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

// GETTKNSTR -- Must return a string, even if that string is all numbers.
func (bufState *BufferState) GETTKNSTR() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = bufState.GetTokenString(true)
	return TOKEN, EOL
}

// ---------------------------------------- GetTokenEOL -------------------------------------------

// GetTokenEOL -- returns the rest of the original string as a string, and returns the EOL indicator.
func (bufState *BufferState) GetTokenEOL(UpperCase bool) (TOKEN TokenType, EOL bool) {
	// GET ToKeN to EndOfLine.
	// This will build a token that consists of every character left on the line.
	// That is, it only stops at the end of line.
	// The TRIM procedure is used to set the COUNT and LENGTH fields.  This is
	// the only TOKENIZE procedure that uses it.

	var Char CharType
	bufState.PREVPOSN = bufState.CURPOSN // So this tkn can be ungotten as well
	tokenRuneSlice := make([]rune, 0, 200)
	TOKEN = TokenType{}
	for {
		Char, EOL = bufState.GetChar() // the Cap function is not here anymore.
		if EOL {
			break
		} //  No-go.  Need to change this to idiomatic Go, but after debugging the main GetTkn and UnGetTkn rtns.
		tokenRuneSlice = append(tokenRuneSlice, Char.Ch)
		//                                                                        TOKEN.Str += string(Char.Ch);
		TOKEN.Isum += int(Char.Ch)
		TOKEN.State = ALLELSE
	}
	TOKEN.Str = string(tokenRuneSlice)

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

// GETTKNEOL -- original token getting rtn, returning a TokenType and EOL indicator.
func (bufState *BufferState) GETTKNEOL() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = bufState.GetTokenEOL(true)
	return TOKEN, EOL
} // GETTKNEOL

//  UNGETTKN is an internal function

// ---------------------------------------- UNGETTKN --------------------------------------------

func (bufState *BufferState) UNGETTKN() {
	/*
	   * UNGET TOKEN ROUTINE.
	   This routine will unget the last token fetched.  It does this by restoring
	   the previous value of POSN, held in PREVPOSN.  Only the last token fetched
	   can be ungotten, so PREVPOSN is reset after use.  If PREVPOSN contains this
	   as its value, then the unget operation will fail.
	*/
	if (bufState.CURPOSN <= bufState.PREVPOSN) || (bufState.PREVPOSN < 0) {
		log.SetFlags(log.Llongfile)
		log.Print("CurPosn out_of_range in UnGetTkn")
		os.Exit(1)
	} // End error trap

	bufState.CURPOSN = bufState.PREVPOSN
	bufState.PREVPOSN = 0
	bufState.strReader1 = bufState.strReader2

	_, err := bufState.strReader1.Seek(int64(bufState.CURPOSN), io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}
}

//                                        GetTokenSlice, now TokenSlice

// TokenSlice -- Single routine that returns a slice of tokens as determined by GetToken.
func TokenSlice(str string) []TokenType {
	if str == "" {
		return nil
	}
	bs := New(str)                        // bs is a buffer state
	tknslice := make([]TokenType, 0, 100) // arbitrary limit, i.e., a magic number as per Rob Pike.

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

// RealTokenSlice -- Single routine that returns a slice of tokens as determined by GETTKNREAL.
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

// TokenRealSlice -- Single routine that returns a slice of tokens as determined by TokenReal.
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

// end tknptr2
