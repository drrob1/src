package main

import (
	"fmt"
	"src/poetry"
)

func main() {
	p, err := poetry.LoadPoem("shortpoem.txt")
	if err != nil {
		fmt.Println(" Error reading from shortpoem.txt", err)
	}
	fmt.Println(p)
	fmt.Println()
	fmt.Printf("%#v\n", p)
}
