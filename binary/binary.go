// import "encoding/binary"

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func main() {
	var pi float64
	b := []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.LittleEndian, &pi)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	fmt.Print(pi)
}




package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func main() {
	buf := new(bytes.Buffer)
	var pi float64 = math.Pi
	err := binary.Write(buf, binary.LittleEndian, pi)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	fmt.Printf("% x", buf.Bytes())
}




package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func main() {
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(61374),
		int8(-54),
		uint8(254),
	}
	for _, v := range data {
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
	}
	fmt.Printf("%x", buf.Bytes())
}




  var err error;
  StackFileExists := true;

  InputByteSlice := make([]byte,8*hpcalc.StackSize);  // I hope this is a slice of 64 bytes, ie, 8*8.
  if InputByteSlice, err = ioutil.ReadFile(StackFileName); err != nil {
    fmt.Errorf(" Error from ioutil.ReadFile.  Probably because no Stack File found: %v\n", err);
    StackFileExists = false;
  }
  if StackFileExists {  // i'll read all into memory.  I just have to lookup how
    for i := 0; i < hpcalc.StackSize*8; i=i+8 {
//      tempByteSlice := InputByteSlice[i:i+8]; and then use tempByteSlice in the NewReader statement that's next.
      buf := bytes.NewReader(InputByteSlice[i:i+8]);
      err := binary.Read(buf,binary.LittleEndian, &R);
      if err != nil {
        fmt.Errorf(" binary.Read failed with error of %v \n",err);
        StackFileExists = false;
      }
      hpcalc.PUSHX(R);
    }  // loop to extract each 8 byte chunk to convert to a longreal (float64) and push onto the hpcalc stack.
  } // stackfileexists

