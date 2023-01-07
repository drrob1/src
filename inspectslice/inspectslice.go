// Here on Win10, running this Jan 7, 2023 shows continguous addresses separated by 0x10, or 16 bytes.  That's 2 words.  But a string is a 3 word struct, ptr, len, cap.
// I'm missing something so far.  Are the len and cap 4 bytes each?  I expect that a pointer is 8 bytes on this 64 bit version of Windows 10.

package main

import "fmt"

func main() {
	fruits := []string{"banana", "apple", "pear", "grape", "peach"}
	inspectSlice(fruits)
}

func inspectSlice(slice []string) {
	fmt.Printf("Length[%d]  Capacity[%d]\n", len(slice), cap(slice))

	for i, s := range slice {
		fmt.Printf(" [%d] %p %s\n", i, &slice[i], s)
	}
	fmt.Println()
}
