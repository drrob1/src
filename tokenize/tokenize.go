package tokenize

import (
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

/*
 Copyright (C) 1987-2016  Robert Solomon MD.  All rights reserved.
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
  20 Mar 89 -- Changed GETOPCODE so that if a multicharacter op code is
                invalid, UNGETCHR is used to put the second char back.
   1 Dec 89 -- Made change in GETTKN that was demonstrated to be necessary
                when the code was ported to the VAX.
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
  28 Dec 14 -- Turns out that CentOS c++ does not support -std=c++11, so I have to remove string.front and string.back
                 member functions.
  18 Jan 15 -- Found bug in that single digits followed by add or subtract are not processed correctly by GETTKNREAL.
  19 Aug 16 -- Finished conversion to Go, started 8/6/16 on boat to Bermuda.
  21 Sep 16 -- Now that this code is for case sensitive filesystem like linux, returning an all caps token is a bad idea.
               So I added FetchToken which takes a param of true for cap and false for preserving case.
   9 Oct 16 -- Will allow "0x" as prefix for hex, as well as "H" suffix.  An 'x' anywhere in the number will
                be a hex number.  I will not force it to be the 2nd character.
  25 Nov 16 -- The TKNMAXSIZ was too small for sha512, so I increased it.
   3 Dec 16 -- Decided to change how the UpperCase flag is handled in GetToken.
  13 Oct 17 -- Made tab char a delim.  Needed for comparehashes.
  14 Oct 17 -- Decided to change the initializing routine so that all control characters are delims.
                 I hope that I don't break anything.  And I'm not changing tknptr package for now.
				 I thought about writing a SetMapDelim and SetMapAllelse, but decided I don't need it for now.
*/

type FSATYP int

const (
	DELIM = iota // so DELIM = 0, and so on.  And the zero val needs to be DELIM.
	OP
	DGT
	ALLELSE
) // FSATYP enum is really int, but will call it FSATYP to indicate intent.

type TokenType struct {
	Str        string
	State      FSATYP
	DelimCH    byte
	DelimState FSATYP
	Isum       int
	Rsum       float64
} // TokenType record

type CharType struct {
	Ch    byte
	State FSATYP
} // CharType Record

// *********************************************************************

const TKNMAXSIZ = 180
const OpMaxSize = 2
const Dgt0 = '0'
const Dgt9 = '9'
const POUNDSIGN = '#' /* 35 */
const PLUSSIGN = '+'  /* 43 */
const COMMA = ','     /* 44 */
const MINUSSIGN = '-' /* 45 */
const SEMCOL = ';'    /* 59 */
const LTSIGN = '<'    /* 60 */
const EQUALSIGN = '=' /* 61 */
const GTSIGN = '>'    /* 62 */
const MULTSIGN = '*'
const DIVSIGN = '/'
const SQUOTE = '\047' /* 39 */
const DQUOTE = '"'    /* 34 */
const NullChar = 0
const EXPSIGN = '^'
const PERCNT = '%'

// These variables are declared here to make the variable global so to maintain their values btwn calls.

var CURPOSN, HOLDCURPOSN, PREVPOSN int

var linebuf, lineByteSlice, HoldLineBS []byte

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

The second argument, quote, specifies the type of literal being parsed and therefore which escaped quote character is permitted. If set to a single quote, it permits the sequence \' and disallows unescaped '. If set to a double quote, it permits \" and disallows unescaped ". If set to zero, it does not permit either escape and allows both quote characters to appear unescaped.
*/

func CAP(c byte) byte {
	if (c >= 'a') && (c <= 'z') {
		c = c - 32
	}
	return c
}

var StateMap map[byte]FSATYP

// ***************************************** init ********************************************
func init() {
	StateMap = make(map[byte]FSATYP, 128)
	InitStateMap()
}

// ************************************ InitStateMap ************************************************
// Making sure that the StateMap is at its default values, since a call to GetTokenStr changes some values.
// Changed initializing routine Oct 14, 2017.  See comments above.
func InitStateMap() {
	//	StateMap[NullChar] = DELIM
	for i := 0; i < 33; i++ {
		StateMap[byte(i)] = DELIM
	}
	for i := 33; i < 128; i++ {
		StateMap[byte(i)] = ALLELSE // including comma
	}
	for c := Dgt0; c <= Dgt9; c++ {
		StateMap[byte(c)] = DGT
	}
	//	StateMap[' '] = DELIM   Not needed anymore.
	//	StateMap['\t'] = DELIM // this is the tab char, but not needed since I changed entire initializing routine
	StateMap[';'] = DELIM
	StateMap['#'] = OP
	StateMap['*'] = OP
	StateMap['+'] = OP
	StateMap['-'] = OP
	StateMap['/'] = OP
	StateMap['<'] = OP
	StateMap['='] = OP
	StateMap['>'] = OP
	StateMap['%'] = OP
	StateMap['^'] = OP
} // InitStateMap

// ***************************************** INITKN *******************************************
func INITKN(Str string) {
	// INITIALIZE TOKEN.
	// The purpose of the initialize token routine is to initialize the
	// variables used by nxtchr to begin processing a new line.
	// The buffer on which the tokenizing rtns operate is also initialized.
	// CURPOSN is initialized to start at the first character on the line.

	InitStateMap() // It's possible that GetTknStr or GetTknEOL changed the StateMap, so will call init.
	CURPOSN = 0
	lineByteSlice = []byte(Str) // do I need to use copy(lineByteSlice,Str) here?  I will see when I get home.
	PREVPOSN = 0
	copy(HoldLineBS, lineByteSlice) // make sure that a value is copied, not just a pointer.
	HOLDCURPOSN = 0
	/*
	   fmt.Println(" In IniTkn and Str is:",Str,", len of Str is: ",len(Str));
	   fmt.Print(" In IniTkn and linebyteslice is '",lineByteSlice);
	   fmt.Println("', length of lineByteSlice is ",len(lineByteSlice));
	   fmt.Println(" In IniTkn and Str is '",Str,"', length= ",len(Str));
	*/
} // INITKN

//****************************** STOTKNPOSN ***********************************
func STOTKNPOSN() {
	/*
	   STORE TOKEN POSITION.
	   This routine will store the value of the curposn into a hold variable for
	   later recall by RCLTKNPOSN.
	*/

	if (CURPOSN < 0) || (CURPOSN > len(lineByteSlice)) {
		log.SetFlags(log.Llongfile)
		log.Print(" In StoTknPosn and CurPosn is invalid.")
		os.Exit(1)
	}
	HOLDCURPOSN = CURPOSN
	copy(HoldLineBS, lineByteSlice) // Need to use copy rtn to copy values, else just copy a pointer so there is no 2nd copy.
} // STOTKNPOSN

//****************************** RCLTKNPOSN **********************************
func RCLTKNPOSN() {
	/*
	   RECALL TOKEN POSITION.
	   THIS IS THE INVERSE OF THE STOTKNPOSN PROCEDURE.
	*/

	if (HOLDCURPOSN < 0) || (len(HoldLineBS) == 0) || (HOLDCURPOSN > len(HoldLineBS)) {
		log.SetFlags(log.Llongfile)
		log.Print(" In RclTknPosn and HoldCurPosn is invalid.")
	}
	CURPOSN = HOLDCURPOSN
	//  lineByteSlice = HoldLineBS;  I think this just copies the references to the slices, not the contents.
} // RCLTKNPOSN

// **************************** PeekChr *******************************************
func PeekChr() (C CharType, EOL bool) {
	/*
	   -- This is the GET CHARACTER ROUTINE.  Its purpose is to get the next
	   -- character from inbuf, determine its fsatyp (finite state automata type),
	   -- and return the upper case value of char.
	   -- NOTE: the curposn pointer is used before it's incremented, unlike most of the
	   -- other pointers in this program.
	      As of 21 Sep 16, the CAP function was removed from here, and conditionally placed into GetToken.
	*/
	EOL = false
	if CURPOSN >= len(lineByteSlice) {
		EOL = true
		C = CharType{} // This zeros the CharType C by assigning an empty CharType constant literal.
		//               C.Ch = NullChar;
		//               C.State = DELIM;
		return C, EOL
	}
	C.Ch = lineByteSlice[CURPOSN] // element access statement
	//                                            C.Ch = CAP((lineByteSlice[CURPOSN]));
	C.State = StateMap[C.Ch] // state assignment, here using map access.
	return C, EOL
} // PeekCHR

// **************************** NextChr  *******************************************
func NextChr() {
	CURPOSN++
} // NextChr

// ********************************* GetChr ********************************
func GETCHR() (C CharType, EOL bool) {
	C, EOL = PeekChr()
	NextChr()
	return C, EOL
} // GETCHR

// ********************************* UNGETCHR ********************************
func UNGETCHR() {
	/*
	   -- UNGETCHaracteR.
	   -- This is the routine that will allow the character last read to be read
	   -- again by decrementing the pointer into the line buffer, CURPOSN.
	*/

	if CURPOSN < 0 {
		log.SetFlags(log.Llongfile)
		log.Print(" CURPOSN out of range in UnGetChr")
		os.Exit(1)
	}
	CURPOSN--
} // UNGETCHR

// ************************************* GetOpCode *********************************************
func GETOPCODE(Token TokenType) int {
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

	if (len(Token.Str) < 1) || (len(Token.Str) > 2) {
		log.SetFlags(log.Llongfile)
		log.Print(" Token too long error from GetOpCode.")
		return OpCode
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
		UNGETCHR() // unget the 2nd part of the invalid pair.
	} // Length of Token = 1
	return OpCode
} // GETOPCODE

//       ***************************=== GetToken ===**************************************
func GetToken(UpperCase bool) (TOKEN TokenType, EOL bool) {
	var (
		CHAR   CharType
		QUOCHR rune /* Holds the active quote char */
	)
	QUOCHR = NullChar
	PREVPOSN = CURPOSN
	TOKEN = TokenType{} // This will zero out all the fields by using a nil struct literal.
	NEGATV := false
	QUOFLG := false

	tokenByteSlice := make([]byte, 0, 200) // to build up the TOKEN.Str field

ExitForLoop:
	for {
		CHAR, EOL = GETCHR()
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
				UNGETCHR()        // To allow correct processing of op pair that is not a valid op, like +- or =>
				break ExitForLoop //goto ExitLoop;
			case OP: // OP -> OP means another operator character found.
				if len(tokenByteSlice) > OpMaxSize {
					UNGETCHR()
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
						TOKEN.Isum = int(CHAR.Ch) - Dgt0
					} else { // TOKEN length > 1 so must first return valid OP
						UNGETCHR()                                   /* UNGET THIS DIGIT CHAR */
						UNGETCHR()                                   /* THEN UNGET THE ARITH SIGN CHAR */
						CHAR.Ch = LastChar                           // SO DELIMCH CORRECTLY RETURNS THE ARITH SIGN CHAR
						tokenByteSlice = tokenByteSlice[:upperbound] // recall that upperbound is excluded in this syntax.
						//                  TOKEN.Str = TOKEN.Str[:upperbound]; // del last char of the token which is the sign character
						break ExitForLoop //goto ExitLoop;
					} // if length of the token = 1
				} else { // IF have a sign character as the lastchar
					UNGETCHR()
					break ExitForLoop //goto ExitLoop;
				} // If have a sign character as the lastchar
			case ALLELSE: // OP -> AllElse
				UNGETCHR()
				break ExitForLoop //goto ExitLoop;
			} // Char.State
		case DGT: // tokenstate
			switch CHAR.State {
			case DELIM:
				break ExitForLoop //goto ExitLoop;
			case OP: // DGT -> OP
				UNGETCHR()
				break ExitForLoop //goto ExitLoop;
			case DGT: // DGT -> DGT so we have another digit.
				tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
				TOKEN.Isum = 10*TOKEN.Isum + int(CHAR.Ch) - Dgt0
			case ALLELSE: // DGT -> AllElse
				UNGETCHR()
				break ExitForLoop //goto ExitLoop;
			} // Char.State
		case ALLELSE: // tokenstate
			switch CHAR.State {
			case DELIM:
				//  Always exit if get a NULL char as a delim.  A quoted string can only get here if CH is NULL.
				break ExitForLoop //goto ExitLoop;
			case OP:
				UNGETCHR()
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
					//                                                                               goto ExitLoop;
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

	//  ExitLoop:
	//  TOKEN.Str = string(tokenByteSlice);  // Trying to apply idiomatic Go guidelines to use byte slice intermediate.

	if UpperCase {
		TOKEN.Str = strings.ToUpper(string(tokenByteSlice))
		//                                                                          TOKEN.Str = strings.ToUpper(TOKEN.Str);
	} else {
		TOKEN.Str = string(tokenByteSlice) // Trying to apply idiomatic Go guidelines to use byte slice intermediate.
	}
	TOKEN.DelimCH = CHAR.Ch
	TOKEN.DelimState = CHAR.State
	if (TOKEN.State == DGT) && NEGATV {
		TOKEN.Isum = -TOKEN.Isum
	}

	//  For OP tokens, must return the opcode as the sum value.  Do this by calling GETOPCODE.
	if TOKEN.State == OP {
		TOKEN.Isum = GETOPCODE(TOKEN)
	}
	TOKEN.Rsum = float64(TOKEN.Isum)
	return TOKEN, EOL
} // GetToken

//------------------------------*************************** GETTKN **************************************
func GETTKN() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = GetToken(true)
	return TOKEN, EOL
} // GETTKN

//********************************** isdigit ***********************************************
func isdigit(ch rune) bool {
	isdgt := ch >= Dgt0 && ch <= Dgt9
	return isdgt
}

//********************************** ishexdigit *************************************************
func ishexdigit(ch rune) bool {

	ishex := isdigit(ch) || ((ch >= 'A') && (ch <= 'F'))
	return ishex

} // ishexdigit

//*********************************** fromhex *************************************************
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

//-------------------------------------------- GETTKNREAL ---------------------------------------
// I am copying the working code from TKNRTNS here.  See the comments in tknrtnsa.adb for reason why.
// Now allows "0x" as hex prefix, as well as "H' as hex suffix.

func GETTKNREAL() (TOKEN TokenType, EOL bool) {
	var CHAR CharType

	TOKEN, EOL = GETTKN()
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
	UNGETTKN()
	TOKEN = TokenType{} // assign nil struct literal to zero all of the fields.
	TOKEN.State = DGT
	PREVPOSN = CURPOSN
	HexFlag := false

ExitLoop:
	for {
		CHAR, EOL = GETCHR()
		CHAR.Ch = CAP(CHAR.Ch)
		if EOL {
			//                        If TKNSTATE is DELIM, then GETTKN was called when there were
			//                        no more tokens on line.  Otherwise it means that we have fetched the last
			//                        token on this line.
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
				UNGETCHR()
				break ExitLoop // goto ExitLoop;
			}
			tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
		case DGT:
			tokenByteSlice = append(tokenByteSlice, CHAR.Ch)
		case ALLELSE:
			if (CHAR.Ch != '.') && (CHAR.Ch != 'E') && !ishexdigit(rune(CHAR.Ch)) && (CHAR.Ch != 'H') &&
				(CHAR.Ch != 'X') {
				UNGETCHR()
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

// *************************************** GetTokenString ***************************************

func GetTokenString(UpperCase bool) (TOKEN TokenType, EOL bool) {
	var Char CharType
	for c := Dgt0; c <= Dgt9; c++ {
		StateMap[byte(c)] = ALLELSE
	}

	StateMap['#'] = ALLELSE /* poundsign */
	StateMap['*'] = ALLELSE /* multsign */
	StateMap['+'] = ALLELSE /* plussign */
	StateMap['-'] = ALLELSE /* minussign */
	StateMap['/'] = ALLELSE /* divsign */
	StateMap['<'] = ALLELSE /* LTSIGN */
	StateMap['='] = ALLELSE /* EQUAL */
	StateMap['>'] = ALLELSE /* GTSIGN */

	TOKEN, EOL = GetToken(UpperCase)
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
		Char, EOL = GETCHR() // the CAP function is not here anymore.
		if EOL || ((Char.State == DELIM) && (len(TOKEN.Str) > 0)) {
			break // Ignore leading delims
		}
		//  NEED TO FIX THIS LATER !!!!!!                                       TOKEN.Str += Char.Ch;  // Here I will leave the code as is and not use a byte slice.
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

// *************************************** GETTKNSTR ***************************************

func GETTKNSTR() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = GetTokenString(true)
	return TOKEN, EOL
}

// **************************************** GetTokenEOL *******************************************
func GetTokenEOL(UpperCase bool) (TOKEN TokenType, EOL bool) {
	// GET ToKeN to EndOfLine.
	// This will build a token that consists of every character left on the line.
	// That is, it only stops at the end of line.
	// The TRIM procedure is used to set the COUNT and LENGTH fields.  This is
	// the only TOKENIZE procedure that uses it.

	var Char CharType
	PREVPOSN = CURPOSN /* So this tkn can be ungotten as well */
	tokenByteSlice := make([]byte, 0, 200)
	TOKEN = TokenType{}
	for {
		Char, EOL = GETCHR() // the Cap function is not here anymore.
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

// ***************************************** GETTKNEOL ******************************************
func GETTKNEOL() (TOKEN TokenType, EOL bool) {
	TOKEN, EOL = GetTokenEOL(true)
	return TOKEN, EOL
} // GETTKNEOL

//************************************** UNGETTKN *****************************
func UNGETTKN() {
	/*
	   * UNGET TOKEN ROUTINE.
	   This routine will unget the last token fetched.  It does this by restoring
	   the previous value of POSN, held in PREVPOSN.  Only the last token fetched
	   can be ungotten, so PREVPOSN is reset after use.  If PREVPOSN contains this
	   as its value, then the unget operation will fail.
	*/
	if (CURPOSN <= PREVPOSN) || (PREVPOSN < 0) {
		log.SetFlags(log.Llongfile)
		log.Print("CurPosn out_of_range in UnGetTkn")
		os.Exit(1)
	} // End error trap

	CURPOSN = PREVPOSN
	PREVPOSN = 0
}

/* A way to output program file and line numbers in an error message.  Must be a closure, I think.  But this is
* cumbersome.  I'll leave the code here in case I figure out a way to make it less cumbersome.
   where := func() {
      _, file, line, _ := runtime.Caller(1)
      log.Fatalf(" In UNGETCHR and CurPosn is < 0.  %s:%d\n", file, line)
   }
  where();
*/
