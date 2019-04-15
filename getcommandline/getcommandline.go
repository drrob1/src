/*
 *
 *  Revision History
 *  ================
 *  27 Oct 14 -- First created as C++ code
 *  12 July 16 -- Started converting to Go
 *   6 Aug 16 -- Still converting
 *  13 Sep 16 -- Rewrote using idiomatic Go
 */

package getcommandline

import "os"
import "strings"

/*  Not idiomatic Go.  My reading says to not use += operator for strings.
func GetCommandLineString() string {
        if len(os.Args) <= 1 {
          return ""
        }
	s := "";
	for _, str  := range os.Args[1:] {  // Remember that os.Args[0] is exec pgm name
          s += str;
	  s += " ";
        }
        s = strings.TrimSpace(s);   // remove trailing space that is always there.
        return s;
}
*/

func GetCommandLineString() string {
	if len(os.Args) <= 1 {
		return ""
	}
	s0 := make([]string, 0, 20)
	for _, str := range os.Args[1:] { // Remember that os.Args[0] is exec pgm name
		s0 = append(s0, str)
	}
	s := strings.Join(s0, " ")
	return s
}

func GetCommandLineByteSlice() []byte {
	if len(os.Args) <= 1 {
		return []byte{}
	}
	s := GetCommandLineString()
	bs := []byte(s)
	return bs
}
