package main

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	. "github.com/klauspost/cpuid/v2"
	"runtime"
	"strconv"
)

/*
13 Apr 25 -- From chapter 13 of Mastering Go, 4th ed.
15 Apr 25 -- Added commas to the long numbers
16 Apr 25 -- Added comparison of Itoa and FormatInt.
*/

const lastAltered = "Apr 16, 2025"

func main() {
	fmt.Printf(" CPUID last altered: %s, compiled using %s\n", lastAltered, runtime.Version())
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
	casheL2Str := strconv.Itoa(int(CPU.Cache.L2))
	casheL2Str = AddCommas(casheL2Str)
	fmt.Println("L2 Cache:", casheL2Str, "bytes")
	casheL3Str := strconv.Itoa(CPU.Cache.L3)
	casheL3Str = AddCommas(casheL3Str)
	fmt.Println("L3 Cache:", casheL3Str, "bytes")
	hz := CPU.Hz
	if hz == 0 { // for when Windows doesn't report a freq, and I want to debug this.
		hz = 10_500_000_000 // 10.5 GHz
	}
	i2a := strconv.Itoa(int(hz))
	hzStr := strconv.FormatInt(hz, 10)
	if i2a == hzStr {
		ctfmt.Printf(ct.Green, false, " strconv.Itoa == strconv.FormatInt\n")
	} else {
		ctfmt.Printf(ct.Red, false, " strconv.Itoa != strconv.FormatInt\n")
	}
	hzStr = AddCommas(hzStr)
	fmt.Printf("Frequency %s hz\n", hzStr)
	hzColorStr, color := getMagnitudeStringHz(hz)
	ctfmt.Printf(color, false, "Frequency %s\n", hzColorStr)
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

// getMagnitudeString -- Makes big numbers much easier to read
func getMagnitudeStringHz(i int64) (string, ct.Color) {
	var s1 string
	var f float64
	var color ct.Color
	switch {
	case i > 1_000_000_000_000: // 1 trillion, or TB
		f = float64(i) / 1_000_000_000_000
		s1 = fmt.Sprintf("%d THz", f)
		color = ct.Red
	case i > 100_000_000_000: // 100 billion
		f = float64(i) / 1_000_000_000
		s1 = fmt.Sprintf("%.4g GHz", f)
		color = ct.White
	case i > 10_000_000_000: // 10 billion
		f = float64(i) / 1_000_000_000
		s1 = fmt.Sprintf("%.4g GHz", f)
		color = ct.White
	case i > 1_000_000_000: // 1 billion, or GB
		f = float64(i) / 1000000000
		s1 = fmt.Sprintf("%.4g GHz", f)
		color = ct.White
	case i > 100_000_000: // 100 million
		f = float64(i) / 1_000_000
		s1 = fmt.Sprintf("%.4g MHz", f)
		color = ct.Yellow
	case i > 10_000_000: // 10 million
		f = float64(i) / 1_000_000
		s1 = fmt.Sprintf("%.4g MHz", f)
		color = ct.Yellow
	case i > 1_000_000: // 1 million, or MB
		f = float64(i) / 1000000
		s1 = fmt.Sprintf("%.4g MHz", f)
		color = ct.Yellow
	case i > 100_000: // 100 thousand
		f = float64(i) / 1000
		s1 = fmt.Sprintf("%.4g KHz", f)
		color = ct.Cyan
	case i > 10_000: // 10 thousand
		f = float64(i) / 1000
		s1 = fmt.Sprintf("%.4g KHz", f)
		color = ct.Cyan
	case i > 1000: // KB
		f = float64(i) / 1000
		s1 = fmt.Sprintf("%.3g KHz", f)
		color = ct.Cyan
	default:
		s1 = fmt.Sprintf("%3d bytes", i)
		color = ct.Green
	}
	return s1, color
}
