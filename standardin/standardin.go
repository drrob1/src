package main

import "fmt"
import "os"
import terminal "github.com/carmark/pseudo-terminal-go/terminal"

// 11-27-17 -- First version, based on a posting to golang-nuts
//               It seems that it still needs a <cr> to continue.  Not exactly what I'm
//               looking for.  Maybe I'm looking for raw mode.

func main() {
	fmt.Println("vim-go")
	oldstate, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(0, oldstate)

	data := make([]byte, 4)
	for i := 0; i < 2; i++ {
		fmt.Print(" input at most 2 bytes:")
		n, err := os.Stdin.Read(data)
		fmt.Printf(" Read %d bytes, %q  with error of %v \n ", n, data, err)
		fmt.Printf(" read data %v -- %#v \n", data, data)
	}
}
