package main;

import (
"os"
"bufio"
"fmt"
//                                                      "strings"
"path/filepath"
"getcommandline"
)



func main() {

  if len(os.Args) <= 1 {
    fmt.Println(" Usage: filestat <FileName>");
    os.Exit(0);
  }
  inbuf := getcommandline.GetCommandLineString();
  filename := filepath.Clean(inbuf);
  FI,err := os.Lstat(filename);

  if err == nil {
    fmt.Println(" Lstat succeeded.  filename: ",FI.Name(),", filesize: ",FI.Size(),".");
  }else { // err != nil
    fmt.Println(" Lstat failed on filename ",filename,".  Error string is ",err.Error(),".");
  }
  fmt.Println();

  FI,err = os.Stat(filename);

  if err == nil {
    fmt.Println(" Stat succeeded.  filename: ",FI.Name(),", filesize: ",FI.Size(),".");
  }else{  // err != nil
    fmt.Println(" Stat failed on filename ",filename,".  Error string is ",err.Error(),".");
  }



  scanner := bufio.NewScanner(os.Stdin)
  for {
    fmt.Print(" Enter filename: ");
    scanner.Scan();
    inbuf = scanner.Text();
    if err := scanner.Err(); err != nil {  // note that this err is a shadow
      fmt.Fprintln(os.Stderr, "reading standard input:", err)
      os.Exit(1);
    }
    if len(inbuf) == 0 {
      os.Exit(0);
    }
    filename = filepath.Clean(inbuf);
    FI,err = os.Lstat(filename);

    if err == nil {
      fmt.Println(" Lstat succeeded.  filename: ",FI.Name(),", filesize: ",FI.Size(),".");
    }else { // err != nil
      fmt.Println(" Lstat failed on filename ",filename,".  Error string is ",err.Error(),".");
    }
    fmt.Println();

    FI,err = os.Stat(filename);

    if err == nil {
      fmt.Println(" Stat succeeded.  filename: ",FI.Name(),", filesize: ",FI.Size(),".");
    }else{  // err != nil
      fmt.Println(" Stat failed on filename ",filename,".  Error string is ",err.Error(),".");
    }
  }
}



/*
from package os
type FileInfo

A FileInfo describes a file and is returned by Stat and Lstat.

type FileInfo interface {
        Name() string       // base name of the file
        Size() int64        // length in bytes for regular files; system-dependent for others
        Mode() FileMode     // file mode bits
        ModTime() time.Time // modification time
        IsDir() bool        // abbreviation for Mode().IsDir()
        Sys() interface{}   // underlying data source (can return nil)
}

func Lstat
func Lstat(name string) (FileInfo, error)

Lstat returns a FileInfo describing the named file. If the file is a symbolic link, the returned FileInfo describes the symbolic link. Lstat makes no attempt to follow the link. If there is an 
error, it will be of type *PathError.


func Stat
func Stat(name string) (FileInfo, error)


type PathError -> PathError records an error and the operation and file path that caused it.

type PathError struct {
        Op   string
        Path string
        Err  error
}

func (*PathError) Error

func (e *PathError) Error() string
*/
