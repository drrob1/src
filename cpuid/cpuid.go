package main

import (
	"fmt"
	. "github.com/klauspost/cpuid/v2"
	"strconv"
)

/*
13 Apr 25 -- From chapter 13 of Mastering Go, 4th ed.
15 Apr 25 -- Added commas to the long numbers
*/

func main() {
	// Print basic CPU information:
	fmt.Println("Name:", CPU.BrandName)
	fmt.Println("PhysicalCores:", CPU.PhysicalCores)
	fmt.Println("LogicalCores:", CPU.LogicalCores)
	fmt.Println("ThreadsPerCore:", CPU.ThreadsPerCore)

	fmt.Println("Family", CPU.Family, "Model:", CPU.Model, "Vendor ID:", CPU.VendorID)
	fmt.Printf(" There are %d features\n", len(CPU.FeatureSet())) // there are ~100 features on Win11 Desktop.
	//fmt.Println("Features:", strings.Join(CPU.FeatureSet(), ","))
	for i, feature := range CPU.FeatureSet() {
		fmt.Print(feature, " ")
		if i%25 == 24 {
			fmt.Println()
		}
	}
	fmt.Println()

	fmt.Println("Cacheline bytes:", CPU.CacheLine)
	fmt.Println("L1 Data Cache:", CPU.Cache.L1D, "bytes")
	fmt.Println("L1 Instruction Cache:", CPU.Cache.L1I, "bytes")
	fmt.Println("L2 Cache:", CPU.Cache.L2, "bytes")
	fmt.Println("L3 Cache:", CPU.Cache.L3, "bytes")
	hzStr := strconv.FormatInt(CPU.Hz, 10)
	hzStr = AddCommas(hzStr)
	fmt.Printf("Frequency %s hz\n", hzStr)
}

// InsertIntoByteSlice -- insert a byte into a slice at a designated position.  Intended to insert a comma into a number string.
func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

// AddCommas -- add comas to a number string.
func AddCommas(instr string) string {
	Comma := []byte{','}

	//BS := make([]byte, 0, 15)
	//BS = append(BS, instr...)
	BS := []byte(instr)

	i := len(BS)

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
}
