package makesubst

import (
	"strings"
)

/* (C) 2016.  Robert W Solomon.  All rights reserved.
REVISION HISTORY
----------------
26 Aug 16 -- First version, in Go before possibly backported to the earlier versions.
 1 Oct 21 -- I just noticed that strings package has a replacer type, that does this.  I'm going to try using that.
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
/*
  The first version of this routine used a ByteSlice.  Then I read an example in Go in 21st Century that uses a
slice of runes, which maked more sense to me.  So I changed from inByteSlice that I called BS, to inRune
which is a slice of runes.  That works with no conversion to byte needed, as s is a rune and single quoted
characters are runes.
*/

func MakeReplaced(instr string) string {
	rplcd := strings.NewReplacer("=", "+", ";", "*")
	return rplcd.Replace(instr)
}
