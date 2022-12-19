package list

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

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
