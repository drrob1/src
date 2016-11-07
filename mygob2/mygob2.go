// I'm going to play with the gob file I/O format.  gob=Go Binary.  I don't plan to change rpng code,
// but I would like to see how I could use gob file I/O format.

package main;


import (
"os"
"fmt"
"encoding/gob"
"runtime"
)

const (
       X = iota  // StackRegNames as int.  No need for a separate type.
       Y
       Z
       T5
       T4
       T3
       T2
       T1
       StackSize
)

type StackType [StackSize]float64;

const StackFilename = "hpstack.sav";  // I don't want these names to clash with rpng.go filenames.
const StorageFilename = "hpstorage.sav"

func main() {
  var inStk,outStk StackType;
  var inStorage,outStorage [36]float64;
  var err error;
  var HomeDir string;

  if runtime.GOOS == "linux" {
    HomeDir = os.Getenv("HOME");
  }else if runtime.GOOS == "windows" {
    HomeDir = os.Getenv("userprofile");
  }else{    // then HomeDir will be empty.
    fmt.Println(" runtime.GOOS does not say linux or windows.  Don't know why.");
  }
  fmt.Println();
  fmt.Println();
  fmt.Println(" GOOS =",runtime.GOOS,".  HomeDir =",HomeDir,".  ARCH=",runtime.GOARCH);
  fmt.Println();
  fmt.Println();


// For this exercise, I have to write the files first.  In rpng.go, I read them first and wrote them
// out upon exit.  
  R := 59.0;
  for i := range outStk {
    outStk[i] = R;
    R *= 2;
  }

  for i := range outStorage {
    outStorage[i] = R;
    R /= 1.5;
  }

  thefile, err := os.Create(StorageFilename);
  check(err);
  defer thefile.Close();
  encoder := gob.NewEncoder(thefile);
  err = encoder.Encode(outStk);
  check(err);
  err = encoder.Encode(outStorage);
  check(err);
//      outfile.Close();

// Time to read and check them.  Note that out the stack and storage were written to the same file.

  thefile.Close();
  thefile, err = os.Open(StorageFilename);
  defer thefile.Close();
//  offset,err := thefile.Seek(0,0); // reset file pointer to the beginning of file.
  check(err);
//  fmt.Println(" New offset for the file reads is ",offset,".  This should be zero.");
  decoder := gob.NewDecoder(thefile);
  err = decoder.Decode(&inStk);
  check(err);
  err = decoder.Decode(&inStorage);

  for i := range inStk {
    if inStk[i] != outStk[i] {
      fmt.Println(" inStk and outStk not equal at element",i,". inStk=",inStk[i],", outStk=",outStk[i]);
    }
  }

  for i := range inStorage {
    if inStorage[i] != outStorage[i] {
      fmt.Println(" inStorage and outStorage != at element",i,". inStor=",inStorage[i],", out=",outStorage[i]);
    }
  }

  fmt.Print(" inStk:");
  fmt.Println(inStk);
  fmt.Println();
  fmt.Print("outStk:")
  fmt.Println(outStk);
  fmt.Println();
  fmt.Println();
  fmt.Println();
  fmt.Print(" inStorage:");
  fmt.Println(inStorage);
  fmt.Println();
  fmt.Print("outStorage:");
  fmt.Println(outStorage);
}

func check(err error) {
  if err != nil {
    panic(err);
  }
}
