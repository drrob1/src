package misc

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

/* (C) 2016-2023.  Robert W Solomon.  All rights reserved.
REVISION HISTORY
----------------
26 Aug 16 -- First version, in Go before possibly backported to the earlier versions.
 1 Oct 21 -- I just noticed that strings package has a replacer type, that does this.  I'm going to try using that.
30 Jun 23 -- Back in the 80's when I was using Modula-2, there was a module called MiscM2, and then StdMiscM2, or something like that.
               I'm doing that now in Go.  This module is now called misc, based on makesubst
*/

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

//  The first version of this routine used a ByteSlice.  Then I read an example in Go in 21st Century that uses a
//slice of runes, which made more sense to me.  So I changed from inByteSlice that I called BS, to inRune
//which is a slice of runes.  That works with no conversion to byte needed, as s is a rune and single quoted
//characters are runes.

func MakeReplaced(instr string) string {
	rplcd := strings.NewReplacer("=", "+", ";", "*")
	return rplcd.Replace(instr)
}

// ----------------------------------------------------- readLine ------------------------------------------------------
// Needed as a bytes reader does not have a readString method.
// This includes the fix so that EOF can only be returned after the last line is returned, even if that line does not end in a newline.

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
