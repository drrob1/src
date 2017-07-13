/*
  REVISION HISTORY
  ----------------
  3/12/15 From modula2.org tutorial.  This section deals with dynamic memory usage.  I will change it around so I better understand it.
              And I've made it a double linked list.
  3/13/15 Will add output of the pointers so I can compare this with the prev and next field contents.  And I changed the name
              of variable AnEntry to AnEntryPointer.
  3/15/15 Converting to C++, and changing more names to be clearer that they are pointers.
  3/18/15 Removed the char cast to see if it still works.  It does.
  3/18/15 Made AdrToHexStr also a function.
  6/7/16 Converted to Go
  6/13/16 Revised AdrToHexStr, and made Initial field a rune based on input from the GoLang forum.
  7/1/16 Included the 2 different methods I wrote for the AdrToHexStr conversion so they are both here.
  7/13/17 -- Looks like I'm adding a 3rd AdrToHexStr based on the Python mooc I'm taking.
*/

package main

import (
	"fmt"
	"strings"
	"unsafe"
)

type FullName struct {
	PrevP     *FullName
	NextP     *FullName
	FirstName string
	Initial   rune
	LastName  string
} // struct FullName

/******************************************************************************************************/

func AdrToHexStrByMasking(adr unsafe.Pointer) string {

	const hex = "0123456789ABCDEF"
	var buf [16]byte

	L := uint64(uintptr(adr))

	for i := range buf {
		buf[i] = hex[L>>uint(60-i*4)&0xF]
	}

	s := string(buf[:])
	for {
		if s[0] != '0' {
			break
		}
		s = s[1:]
	}

	return s // without the above loop removing leading zeros, this form does return them
	//  return string(buf[:]) ;  // this form does not return leading zeros on Linux.  Don't know why
} // AdrToHexStr            This entire function can be replaced by a call to Sprintf, but nevermind that

/**************************************************************************************************/
// This code is from the Python course.  It's simplicity is very elegant.  And I'm more experienced now.  07/13/2017 05:51:39 PM
func AdrToHexString(adr unsafe.Pointer) string {

	const ASCZERO int64 = '0'
	const hex = "0123456789ABCDEF"
	const ascA int64 = 'A'
	var h int64

	str := ""

	for L := int64(uintptr(adr)); L > 0; L = L / 16 { // repeat  until L = 0
		h = L % 16 // % is MOD op
		str = string(hex[h]) + str
	} // until L = 0

	return str
} // AdrToHexString            This entire function can be replaced by a call to Sprintf, but nevermind that
/**************************************************************************************************/

func AdrToHexStrByStringSlices(adr unsafe.Pointer) (OutStr string) {

	const ASCZERO int64 = '0'
	const ascA int64 = 'A'
	var h int64

	str := make([]string, 16) // 16 hex digits to be filled in reverse
	i := 0

	L := int64(uintptr(adr))

	for { // repeat  until L = 0
		h = L % 16 // % is MOD op
		if h <= 9 {
			str[15-i] = string(h + ASCZERO)
		} else {
			str[15-i] = string(h - 10 + ascA)
		} // if h <= 9
		i++
		L = L / 16
		if L == 0 {
			break
		}
	} // until L = 0

	OutStr = strings.Join(str, "")
	//                                    I don't think I need this anymore          OutStr = strings.Trim(OutStr," ");
	return
} // AdrToHexStrByStringSlices            This entire function can be replaced by a call to Sprintf, but nevermind that

/**************************************************************************************************/
func AdrToHexStrByByteSlices(adr unsafe.Pointer) string {

	const ASCZERO int64 = '0'
	const ascA int64 = 'A'
	const hex = "0123456789ABCDEF"
	var h int64

	buf := make([]byte, 16) // 16 hex digits to be filled in reverse
	i := 0

	L := int64(uintptr(adr))

	for { // repeat  until L = 0
		h = L % 16 // % is MOD op
		buf[15-i] = hex[h]
		//                                  if h <= 9 {
		//                                    str[15-i] = string(h + ASCZERO);
		//                                  }else{
		//                                    str[15-i] = string(h -10 + ascA);
		//                                  }; // if h <= 9
		i++
		L = L / 16
		if L == 0 {
			break
		}
	} // until L = 0

	//                                                                               OutStr = strings.Join(str,"");
	//                                    I don't think I need this anymore          OutStr = strings.Trim(OutStr," ");
	return string(buf[:])
} // AdrToHexStr            This entire function can be replaced by a call to Sprintf, but nevermind that

/**************************************************************************************************/
func main() {

	/*
	   func AdrToHexStrByMasking(adr unsafe.Pointer) string {
	   func AdrToHexStrByStringSlices(adr unsafe.Pointer) (OutStr string) {
	   func AdrToHexStrByByteSlices(adr unsafe.Pointer) string {
	*/
	var (
		StartOfListP, EndofListP, CurrentPlaceInListP, PrevPlaceInListP, AnEntryPointer *FullName
		unsafeP                                                                         unsafe.Pointer
	)

	fmt.Println()
	fmt.Println()
	StartOfListP = nil
	EndofListP = nil
	CurrentPlaceInListP = nil
	PrevPlaceInListP = nil

	/* Generate the first name in the list */
	AnEntryPointer = new(FullName)
	StartOfListP = AnEntryPointer
	fmt.Print(" 1: ")
	unsafeP = unsafe.Pointer(AnEntryPointer)
	s := AdrToHexStrByMasking(unsafeP)
	s2 := AdrToHexStrByByteSlices(unsafeP)
	fmt.Printf("First pointer value %p, %#v\n", AnEntryPointer, AnEntryPointer)
	fmt.Println(" First pointer value as a string:", s, s2)

	AnEntryPointer.PrevP = nil         // do I need to dereference all of these?
	AnEntryPointer.FirstName = "John " // or is this covered by "syntactic sugar"?
	AnEntryPointer.Initial = 'Q'       // Seems it is covered by "syntactic sugar"
	AnEntryPointer.LastName = " Doe"   // The explicit dereferences are not needed.
	AnEntryPointer.NextP = nil

	/* Generate 2nd name in the list */
	PrevPlaceInListP = AnEntryPointer
	AnEntryPointer = new(FullName)
	fmt.Print(" 2: ")
	unsafeP = unsafe.Pointer(AnEntryPointer)
	s = AdrToHexStrByStringSlices(unsafeP)
	s1 := AdrToHexStrByByteSlices(unsafeP)
	fmt.Printf("%p, %#V\n", AnEntryPointer, AnEntryPointer) // using the %p verb format specifier for a pointer
	fmt.Println(" 2nd pointer as a string:", s, s1)
	CurrentPlaceInListP = AnEntryPointer
	(*PrevPlaceInListP).NextP = CurrentPlaceInListP // This explicit dereference is not needed
	(*CurrentPlaceInListP).PrevP = PrevPlaceInListP
	(*CurrentPlaceInListP).FirstName = "Mary "
	(*CurrentPlaceInListP).Initial = 'R'
	(*CurrentPlaceInListP).LastName = " Johnson"
	(*CurrentPlaceInListP).NextP = nil

	/* Add 10 more names to complete the list */
	for I := 1; I <= 10; I++ {
		PrevPlaceInListP = CurrentPlaceInListP
		AnEntryPointer = new(FullName)
		fmt.Print(I+2, ":")
		unsafeP = unsafe.Pointer(AnEntryPointer)
		s = AdrToHexStrByByteSlices(unsafeP)
		fmt.Printf("%s,%#v\n", s, AnEntryPointer)
		if (I % 3) == 0 {
			fmt.Println()
		}
		CurrentPlaceInListP = AnEntryPointer
		PrevPlaceInListP.NextP = CurrentPlaceInListP
		CurrentPlaceInListP.PrevP = PrevPlaceInListP
		CurrentPlaceInListP.FirstName = "Billy "
		CurrentPlaceInListP.Initial = rune(I + 64) // 65 is cap A
		CurrentPlaceInListP.LastName = " Franklin"
		CurrentPlaceInListP.NextP = nil
	} /* for I */
	EndofListP = CurrentPlaceInListP
	fmt.Println()
	fmt.Println()

	/* Display the list on the monitor in forward direction */
	fmt.Println(" List in forward direction.")
	CurrentPlaceInListP = StartOfListP
	for {
		if CurrentPlaceInListP == nil {
			break
		}
		//    s,s0 = AdrToHexStr(CurrentPlaceInListP.PrevP);
		fmt.Printf("%p: ", CurrentPlaceInListP.PrevP)
		fmt.Printf("%s %c %s: ", CurrentPlaceInListP.FirstName, CurrentPlaceInListP.Initial, CurrentPlaceInListP.LastName)
		//    s,s0 = AdrToHexStr(CurrentPlaceInListP.NextP);
		fmt.Printf("%p\n", CurrentPlaceInListP.NextP)
		PrevPlaceInListP = CurrentPlaceInListP
		CurrentPlaceInListP = CurrentPlaceInListP.NextP
	}

	/* Display the list on the monitor in reverse direction */
	fmt.Println()
	fmt.Println(" List in reverse direction. ")
	CurrentPlaceInListP = EndofListP
	for {
		if CurrentPlaceInListP == nil {
			break
		}
		s0 := AdrToHexString(unsafe.Pointer(CurrentPlaceInListP.PrevP)) // need to pass the pointer in to the conversion routine.
		fmt.Printf("%s  %p: ", s0, CurrentPlaceInListP.PrevP)
		fmt.Printf("%s %c %s: ", CurrentPlaceInListP.FirstName, CurrentPlaceInListP.Initial, CurrentPlaceInListP.LastName)
		s := AdrToHexString(unsafe.Pointer(CurrentPlaceInListP.NextP))
		fmt.Printf("%s  %p\n", s, CurrentPlaceInListP.NextP)
		PrevPlaceInListP = CurrentPlaceInListP
		CurrentPlaceInListP = CurrentPlaceInListP.PrevP
	}

	/* Deallocate is unnecessary in Go */

} // LinkList
/*
func AdrToHexStrByMasking(adr unsafe.Pointer) string {
func AdrToHexStrByStringSlices(adr unsafe.Pointer) (OutStr string) {
func AdrToHexStrByByteSlices(adr unsafe.Pointer) string {
*/
