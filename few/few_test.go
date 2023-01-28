package few_test

import (
	"bufio"
	"fmt"
	"os"
	"src/few"
	"testing"
)

func TestFeq1(t *testing.T) {
	f1, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt 2nd time is %s\n", err)
		return
	}

	b1 := bufio.NewReader(f1)
	b2 := bufio.NewReader(f2)

	if few.Feq1(b1, b2) {
		fmt.Printf(" Success from feq1 for PaulKrugman    ")
		t.Log(" Feq1 for PaulKrugman.txt succeeded")
	} else {
		t.Errorf(" Expected to succeed Feq1 for PaulKrugman but it failed.\n")
		return
	}
	f1.Close()
	f2.Close()

	f1, err = os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from second opening f1= PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err = os.Open("testdata/qpid.txt")
	if err != nil {
		t.Errorf(" Error from opening qpid.txt is %s\n", err)
		return
	}

	b1 = bufio.NewReader(f1)
	b2 = bufio.NewReader(f2)

	if few.Feq1(b1, b2) {
		t.Errorf(" Success from feq1 for PaulKrugman and qpid.txt.  Should have failed\n")
		return
	} else {
		fmt.Printf(" Expected to fail Feq1 for PaulKrugman and qpid, and did.\n")
	}
	f1.Close()
	f2.Close()
	fmt.Printf("\n\n")
}

func TestFeq2(t *testing.T) {
	f1, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt 2nd time is %s\n", err)
		return
	}

	b1 := bufio.NewReader(f1)
	b2 := bufio.NewReader(f2)

	if few.Feq2(b1, b2) {
		fmt.Printf(" Success from feq2 for PaulKrugman    ")
		t.Log(" Feq2 for PaulKrugman.txt succeeded")
	} else {
		t.Errorf(" Expected to succeed Feq2 for PaulKrugman but it failed.\n")
		return
	}
	f1.Close()
	f2.Close()

	f1, err = os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from second opening f1= PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err = os.Open("testdata/qpid.txt")
	if err != nil {
		t.Errorf(" Error from opening qpid.txt is %s\n", err)
		return
	}

	b1 = bufio.NewReader(f1)
	b2 = bufio.NewReader(f2)

	if few.Feq2(b1, b2) {
		t.Errorf(" Success from feq2 for PaulKrugman and qpid.txt.  Should have failed\n")
		return
	} else {
		fmt.Printf(" Expected to fail Feq2 for PaulKrugman and qpid, and did.\n")
	}
	f1.Close()
	f2.Close()

	fmt.Printf("\n\n")

}

func TestFeq32(t *testing.T) {
	f1, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt 2nd time is %s\n", err)
		return
	}

	b1 := bufio.NewReader(f1)
	b2 := bufio.NewReader(f2)

	if few.Feq32(b1, b2) {
		fmt.Printf(" Success from feq32 for PaulKrugman\n")
		t.Log(" Feq32 for PaulKrugman.txt succeeded")
	} else {
		t.Errorf(" Expected to succeed Feq32 for PaulKrugman but it failed.\n")
		return
	}
	f1.Close()
	f2.Close()

	f1, err = os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from second opening f1= PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err = os.Open("testdata/qpid.txt")
	if err != nil {
		t.Errorf(" Error from opening qpid.txt is %s\n", err)
		return
	}

	b1 = bufio.NewReader(f1)
	b2 = bufio.NewReader(f2)

	if few.Feq32(b1, b2) {
		t.Errorf(" Success from feq32 for PaulKrugman and qpid.txt.  Should have failed\n")
		return
	} else {
		fmt.Printf(" Expected to fail Feq32 for PaulKrugman and qpid, and did.\n")
	}
	f1.Close()
	f2.Close()

	fmt.Printf("\n\n")
}
func TestFeq3(t *testing.T) {
	f1, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt 2nd time is %s\n", err)
		return
	}

	b1 := bufio.NewReader(f1)
	b2 := bufio.NewReader(f2)

	if few.Feq3(b1, b2) {
		fmt.Printf(" Success from feq3 for PaulKrugman    ")
		t.Log(" Feq3 for PaulKrugman.txt succeeded")
	} else {
		t.Errorf(" Expected to succeed Feq3 for PaulKrugman but it failed.\n")
		return
	}
	f1.Close()
	f2.Close()

	f1, err = os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from second opening f1= PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err = os.Open("testdata/qpid.txt")
	if err != nil {
		t.Errorf(" Error from opening qpid.txt is %s\n", err)
		return
	}

	b1 = bufio.NewReader(f1)
	b2 = bufio.NewReader(f2)

	if few.Feq3(b1, b2) {
		t.Errorf(" Success from feq3 for PaulKrugman and qpid.txt.  Should have failed\n")
		return
	} else {
		fmt.Printf(" Expected to fail Feq3 for PaulKrugman and qpid, and did.\n")
	}
	f1.Close()
	f2.Close()

	fmt.Printf("\n\n")
}

func TestFeq5(t *testing.T) {
	f1, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt 2nd time is %s\n", err)
		return
	}

	b1 := bufio.NewReader(f1)
	b2 := bufio.NewReader(f2)

	if few.Feq5(b1, b2) {
		fmt.Printf(" Success from feq5 for PaulKrugman    ")
		t.Log(" Feq5 for PaulKrugman.txt succeeded")
	} else {
		t.Errorf(" Expected to succeed Feq5 for PaulKrugman but it failed.\n")
		return
	}
	f1.Close()
	f2.Close()

	f1, err = os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from second opening f1= PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err = os.Open("testdata/qpid.txt")
	if err != nil {
		t.Errorf(" Error from opening qpid.txt is %s\n", err)
		return
	}

	b1 = bufio.NewReader(f1)
	b2 = bufio.NewReader(f2)

	if few.Feq5(b1, b2) {
		t.Errorf(" Success from feq5 for PaulKrugman and qpid.txt.  Should have failed\n")
		return
	} else {
		fmt.Printf(" Expected to fail Feq5 for PaulKrugman and qpid, and did.\n")
	}
	f1.Close()
	f2.Close()

	fmt.Printf("\n\n")
}

func TestFeq64(t *testing.T) {
	f1, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt 2nd time is %s\n", err)
		return
	}

	b1 := bufio.NewReader(f1)
	b2 := bufio.NewReader(f2)

	if few.Feq64(b1, b2) {
		fmt.Printf(" Success from feq64 for PaulKrugman    ")
		t.Log(" Feq64 for PaulKrugman.txt succeeded")
	} else {
		t.Errorf(" Expected to succeed Feq64 for PaulKrugman but it failed.\n")
		return
	}
	f1.Close()
	f2.Close()

	f1, err = os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from second opening f1= PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err = os.Open("testdata/qpid.txt")
	if err != nil {
		t.Errorf(" Error from opening qpid.txt is %s\n", err)
		return
	}

	b1 = bufio.NewReader(f1)
	b2 = bufio.NewReader(f2)

	if few.Feq64(b1, b2) {
		t.Errorf(" Success from feq64 for PaulKrugman and qpid.txt.  Should have failed\n")
		return
	} else {
		fmt.Printf(" Expected to fail Feq64 for PaulKrugman and qpid, and did.\n")
	}
	f1.Close()
	f2.Close()

	fmt.Printf("\n\n")
}

func TestFeqbbb(t *testing.T) {
	f1, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err := os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from opening PaulKrugman.txt 2nd time is %s\n", err)
		return
	}

	b1 := bufio.NewReader(f1)
	b2 := bufio.NewReader(f2)

	BOOL, err := few.Feqbbb(b1, b2)
	if err != nil {
		t.Errorf(" Error from Feqbbb is %s\n", err)
		return
	}
	if BOOL {
		fmt.Printf(" Success from feqbbb for PaulKrugman    ")
		t.Log(" Feqbbb for PaulKrugman.txt succeeded")
	} else {
		t.Errorf(" Expected to succeed Feqbbb for PaulKrugman but it failed.  Will continue to next test.\n")
	}
	f1.Close()
	f2.Close()

	f1, err = os.Open("testdata/PaulKrugman.txt")
	if err != nil {
		t.Errorf(" Error from second opening f1= PaulKrugman.txt is %s\n", err)
		return
	}

	f2, err = os.Open("testdata/qpid.txt")
	if err != nil {
		t.Errorf(" Error from opening qpid.txt is %s\n", err)
		return
	}

	b1 = bufio.NewReader(f1)
	b2 = bufio.NewReader(f2)

	BOOL, err = few.Feqbbb(b1, b2)
	if err != nil {
		t.Errorf(" Error from Feqbbb is %s\n", err)
		return
	}
	if BOOL {
		t.Errorf(" Success from feqbbb for PaulKrugman and qpid.txt.  Should have failed\n")
		return
	} else {
		fmt.Printf(" Expected to fail Feqbbb for PaulKrugman and qpid, and did.\n")
	}
	f1.Close()
	f2.Close()

	fmt.Printf("\n\n")
}
