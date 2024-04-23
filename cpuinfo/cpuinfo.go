package main

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/sys/cpu"
	"runtime"
)

/*
21 Apr 24 -- First version
*/

const lastAltered = "Apr 21, 2024"

func main() {
	ctfmt.Printf(ct.Yellow, true, "CPU Info, last altered %s, compiled with %s\n\n", lastAltered, runtime.Version())
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	numcpu := runtime.NumCPU()
	ctfmt.Printf(ct.Cyan, true, " GOOS = %s, goarch = %s, numcpu = %d\n", goos, goarch, numcpu)
	//ctfmt.Printf(ct.Yellow, true, " GOOS = %s, goarch = %s, numcpu = %d\n", goos, goarch, numcpu)
	if cpu.IsBigEndian {
		ctfmt.Printf(ct.Green, false, " Is big endian\n")
	} else {
		ctfmt.Printf(ct.Red, true, " Is not big endian\n")
	}
	if goarch == "amd64" {
		fmt.Printf(" HasAES = %t, HasADX add carry extension = %t, HasCX16 compare and exchange 16 bytes = %t \n",
			cpu.X86.HasAES, cpu.X86.HasADX, cpu.X86.HasCX16)
		fmt.Printf(" HasFMA fused multiply add = %t, HasRDRAND on chip rand generator = %t, HasRDSEED on chip rand seed = %t, and HasSSE42 streaming SIMD ext 4 and 4.2 = %t \n",
			cpu.X86.HasFMA, cpu.X86.HasRDRAND, cpu.X86.HasRDSEED, cpu.X86.HasSSE42)
		fmt.Printf(" HasBMI2 bit manip instrn 2 = %t, HasERMS enhanced rep for MOVSB and STOSB= %t \n",
			cpu.X86.HasBMI2, cpu.X86.HasERMS)
		fmt.Printf(" HasAMX Tile advanced matrix extension tile = %t, HasAMXInt8 = %t, HasAMSBF16 BFloat16 = %t, \n",
			cpu.X86.HasAMXTile, cpu.X86.HasAMXInt8, cpu.X86.HasAMXBF16)
		if cpu.X86.HasOSXSAVE {
			ctfmt.Printf(ct.Green, false, " Has OS XSAVE, where OS supports XSAVE/XRESTOR w/ XMM registers.\n")
		} else {
			ctfmt.Printf(ct.Red, true, " Does not have OS XSAVE, where OS supports XSAVE/XRESTOR w/ XMM registers.\n")
		}
		if cpu.X86.HasAVX512 {
			ctfmt.Printf(ct.Green, false, " yes AVX512 advanced vector extension \n")
		} else {
			ctfmt.Printf(ct.Red, true, " no AVX512 advanced vector extension. ")
			if cpu.X86.HasAVX {
				ctfmt.Printf(ct.Green, false, " yes AVX advanced vector extension: ")
				if cpu.X86.HasAVX2 {
					ctfmt.Printf(ct.Green, false, " yes AVX2 advanced vector extension \n")
				} else {
					ctfmt.Printf(ct.Red, false, " no AVX2 advanced vector extension \n")
				}
			} else {
				ctfmt.Printf(ct.Red, false, " no AVX advanced vector extension \n")
			}
		}
	}
	fmt.Println()
}
