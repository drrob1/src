package misc

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"strings"
)

/* (C) 2016-2023.  Robert W Solomon.  All rights reserved.
REVISION HISTORY
----------------
26 Aug 16 -- First version, in Go before possibly backported to the earlier versions.
 1 Oct 21 -- I just noticed that strings package has a replacer type, that does this.  I'm going to try using that.
30 Jun 23 -- Back in the 80's when I was using Modula-2, there was a module called MiscM2, and then StdMiscM2, or something like that.
               I'm doing that now in Go.  This module is now called misc, based on makesubst
24 May 24 -- Adding comments that can be processed by go doc.
 9 Nov 24 -- Added Floor function to automatically correct small floating point errors.  It works in hpcalc2, so I copied it here, too.
24 Nov 24 -- Enhanced Floor function to validate the places param.
*/

// MakeSubst -- input a string, output a string that substitutes '=' -> '+' and ';' -> '*'
func MakeSubst(instr string) string {
	instr = strings.TrimSpace(instr)
	inRune := make([]rune, len(instr)) // was a slice of byte in the 1st version of this routine.

	for i, s := range instr {
		switch s {
		case '=':
			s = '+'
		case ';':
			s = '*'
		}
		inRune[i] = s // was byte(s) before I made this a slice of runes.
	}
	return string(inRune)
} // makesubst

// The first version of this routine used a ByteSlice.  Then I read an example in Go in 21st Century that uses a slice of runes, which made more sense to me.
// So I changed from inByteSlice that I called BS, to inRune which is a slice of runes.
// That works with no conversion to byte needed, as s is a rune and single quoted characters are runes.

// MakeReplaced -- input a string, and uses strings.NewReplacer to '=' -> '+' and ';' -> '*'
func MakeReplaced(instr string) string {
	rplcd := strings.NewReplacer("=", "+", ";", "*")
	return rplcd.Replace(instr)
}

// ----------------------------------------------------- readLine ------------------------------------------------------
// Needed as a bytes reader does not have a readString method.
// This includes the fix so that EOF can only be returned after the last line is returned, even if that line does not end in a newline.

// ReadLine -- input a *bytes.Reader, and outputs a line delimited by \n, and an error returned by ReadByte that is not io.EOF.  io.EOF is expected for each line, so that's not an error.
func ReadLine(r *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		byte, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if sb.Len() > 0 {
					return sb.String(), nil
				}
				// Error here is not EOF.
				return strings.TrimSpace(sb.String()), err
			}
		}
		if byte == '\n' {
			return strings.TrimSpace(sb.String()), nil
		}
		err = sb.WriteByte(byte)
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
	}
} // readLine

// ------------------------------------------------randRange -----------------------------------------------------------
// I learned about this from the manning live project that taught RSA public key cryptography.
//

// RandRange -- Input a min and a max int, and returns a random int in that range.
func RandRange(minP, maxP int) int { // note that this is not cryptographically secure.  Would need crypto/rand for that.
	if maxP < minP {
		minP, maxP = maxP, minP
	}
	return minP + rand.IntN(maxP-minP)
}

// CreateOrAppendWithBuffer This takes a name string and returns a file pointer opened for writing using the bufio routines.  Will not truncate or clobber a file.
func CreateOrAppendWithBuffer(name string) (*os.File, *bufio.Writer, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}
	buf := bufio.NewWriter(f)
	return f, buf, nil
}

// CreateOrAppendWithoutBuffer This take a name string and returns a simple file pointer, NOT using the bufio routines.  Will not truncate or clobber a file.
func CreateOrAppendWithoutBuffer(name string) (*os.File, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	return f, err
}

// AddCommasRune will use runes to add a comma into a numeric string.  Just to see if this works.  It does.
func AddCommasRune(instr string) string {
	//var Comma []byte = []byte{','}  Getting message that type can be omitted.
	comma := []rune{','}

	RS := []rune(instr)

	i := len(RS)

	for numberOfCommas := i / 3; (numberOfCommas > 0) && (i > 3); numberOfCommas-- {
		i -= 3
		RS = InsertIntoRuneSlice(RS, comma, i)
	}
	return string(RS)
} // AddCommasRune

func InsertIntoRuneSlice(slice, insertion []rune, index int) []rune {
	return append(slice[:index], append(insertion, slice[index:]...)...)
} // InsertIntoByteSlice

// Floor -- To automatically fix the small floating point errors introduced by the conversions.  Max value for places is 10.
func Floor(real, places float64) float64 {
	if places < 0 || places > 10 {
		places = 10
	}

	negFlag := real < 0
	result := real
	if negFlag {
		result *= -1
	}
	factor := math.Pow(10, places)
	result *= factor
	result = math.Floor(result + 0.5)
	result /= factor
	if negFlag {
		result *= -1
	}
	return result
}
