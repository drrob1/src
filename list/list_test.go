package list

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func TestExpandADash(t *testing.T) {
	str := "a-c"
	out, err := ExpandADash(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandDash.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "abc" {
		t.Errorf(" out should be abc but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct In=%#v, out=%#v\n", str, out)
	}

	str = "b-d"
	out, err = ExpandADash(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandDash.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "bcd" {
		t.Errorf(" out should be bcd but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct In=%#v, out=%#v\n", str, out)
	}

	str = "s-v"
	out, err = ExpandADash(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandDash.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "stuv" {
		t.Errorf(" out should be xyz but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct In=%#v, out=%#v\n", str, out)
	}

	str = "s-vxyz"
	out, err = ExpandADash(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandDash.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "stuvxyz" {
		t.Errorf(" out should be xyz but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct In=%#v, out=%#v\n", str, out)
	}

	str = "x-z"
	out, err = ExpandADash(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandDash.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "xyz" {
		t.Errorf(" out should be xyz but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct In=%#v, out=%#v\n", str, out)
	}

	str = "abcx-z"
	out, err = ExpandADash(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandDash.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "abcxyz" {
		t.Errorf(" out should be xyz but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct In=%#v, out=%#v\n", str, out)
	}

	str = "abc-quvw"
	out, err = ExpandADash(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandDash.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "abcdefghijklmnopquvw" {
		t.Errorf(" out should be xyz but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct In=%#v, out=%#v\n", str, out)
	}

	str = "a-b"
	out, err = ExpandADash(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandDash.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "ab" {
		t.Errorf(" out should be ab but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct In=%#v, out=%#v\n", str, out)
	}

	str = "abcde"
	out, err = ExpandADash(str)
	if err != nil {
		t.Errorf(" No expansion found.  But should not be an error.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "abcde" {
		t.Errorf(" out should be abcde but is %#v instead.\n", out)
	} else {
		fmt.Printf(" No expansion done.  In=%#v, out=%#v\n", str, out)
	}

	out, err = ExpandADash("a-")
	if err != nil {
		fmt.Printf(" err is No ending character found.  err=%#v and out=%#v, err=%#v\n", err, out, err)
	}

	out, err = ExpandADash("a- b")
	if err != nil {
		fmt.Printf(" err is Invalid index found.  err=%#v and out=%#v\n", err, out)
	}
}

func TestExpandAllDashes(t *testing.T) {
	str := "a-cegqz"
	out, err := ExpandAllDashes(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandAllDashes.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "abcegqz" {
		t.Errorf(" out should be abcegqz but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct: In=%#v, out=%#v\n", str, out)
	}

	str = "b-de-gj-m"
	out, err = ExpandAllDashes(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandAllDashes.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "bcdefgjklm" {
		t.Errorf(" out should be bcdefgjklm but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct: In=%#v, out=%#v\n", str, out)
	}

	str = "ab-de-gj-mpqvz"
	out, err = ExpandAllDashes(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandallDashes.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "abcdefgjklmpqvz" {
		t.Errorf(" out should be abcdefgjklmpqvz but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct: In=%#v, out=%#v\n", str, out)
	}

	str = "x-zabc"
	out, err = ExpandAllDashes(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandallDashes.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "xyzabc" {
		t.Errorf(" out should be xyzabc but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct: In=%#v, out=%#v\n", str, out)
	}

	str = "a-b"
	out, err = ExpandAllDashes(str)
	if err != nil {
		t.Fatalf(" ERROR: from call to ExpandallDashes.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "ab" {
		t.Errorf(" out should be ab but is %#v instead.\n", out)
	} else {
		fmt.Printf(" In=%#v, out=%#v\n", str, out)
	}

	str = "abcde"
	out, err = ExpandAllDashes(str)
	if err != nil {
		t.Errorf(" No expansion found.  But should not be an error for ExpandAllDashes.  str=%#v, out=%#v, err=%#v\n", str, out, err)
	}
	if out != "abcde" {
		t.Errorf(" out should be abcde but is %#v instead.\n", out)
	} else {
		fmt.Printf("Correct: No expansion done.  In=%#v, out=%#v\n", str, out)
	}

	out, err = ExpandAllDashes("a-")
	if err != nil {
		fmt.Printf(" err is not supposed to be nil.  err=%#v and out=%#v, err=%#v\n", err, out, err)
	}

	out, err = ExpandAllDashes("a- b")
	if err != nil {
		fmt.Printf(" err is not supposed to be nil.  err=%#v and out=%#v\n", err, out)
	}
}

func TestIncludeThis(t *testing.T) {
	fi, err := os.Stat("./testdata/file.txt")
	if err != nil {
		t.Fatalf(" ERROR: from first call to os.Stat is %#v\n", err)
	}
	var excludeMe *regexp.Regexp
	excludeMe, _ = regexp.Compile("")
	includeMe := includeThis(fi, excludeMe)
	if includeMe {
		fmt.Printf(" for %s, includeMe is %t, which is correct.\n", fi.Name(), includeMe)
	} else {
		t.Errorf(" for %s, expected includeMe to be true, but it is %t\n", fi.Name(), includeMe)
	}

	fi, err = os.Stat("./testdata/notme.txt")
	excludeMe, err = regexp.Compile("not")
	if err != nil {
		t.Fatalf(" ERROR from regexp.Compile is %s\n", err)
	}
	includeMe = includeThis(fi, excludeMe)
	if includeMe {
		fmt.Errorf(" includeMe for %s should be false, but it is %t\n", fi.Name(), includeMe)
	} else {
		fmt.Printf(" includeMe for %s is false, which is correct\n", fi.Name())
	}

	fi, err = os.Stat("./testdata/menot.txt")
	excludeMe, err = regexp.Compile("not")
	if err != nil {
		t.Fatalf(" ERROR from regexp.Compile is %s\n", err)
	}
	includeMe = includeThis(fi, excludeMe)
	if includeMe {
		fmt.Errorf(" includeMe for %s should be false, but it is %t\n", fi.Name(), includeMe)
	} else {
		fmt.Printf(" includeMe for %s is false, which is correct\n", fi.Name())
	}
}
