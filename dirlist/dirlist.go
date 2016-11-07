package main
import (
  "os"
  "fmt"
  "path/filepath"
  "time"
)

func main() {
  dirname := "." + string(filepath.Separator);
  fmt.Println("dirname is",dirname);
  d, err := os.Open(dirname);
  if err != nil {
    fmt.Println(err);
    os.Exit(1);
  }
  
  defer d.Close();

  fi, err := d.Readdir(-1);
  if err != nil {
    fmt.Println(err);
    os.Exit(1);
  }

  DirEntries := make([]os.FileInfo,len(fi) );   // This made the slice with too many entries.  And append would not fill the earlier empty spots.  Now that I think of it, this may make sense as ...
//  FileNames := make([]string,len(fi) );         This made the slice with too many entries.  I don't know why.
  FileNames := make([]string,0);
  i := 0;
  for _, fi := range fi {   // I think this is an intential shadowing of fi with fi
    if fi.Mode().IsRegular() {
      fmt.Println(i,":",fi.Mode(),fi.ModTime(),fi.ModTime().UnixNano(),fi.Name(),fi.Size(),"bytes");

//      fmt.Printf(" mod time is %v of type %T, and type %T\n",fi.ModTime(),fi.ModTime(),fi.ModTime().UnixNano() ) ;  // The answer is of type time.Time, and type int64
      DirEntries[i] = fi;  // I'm looking to not have the first 2 entries be empty and cause a panic when I tried to access .Name();
      i++;
      FileNames = append(FileNames,fi.Name());
    }
  }
  fmt.Println("Len of Directory Entries slice is",len(DirEntries));
  fmt.Println("Len of regular filenames slice is",len(FileNames));

//                                                       fmt.Println(" Regular Directory Entries in the slice of them are:");
//                                                       fmt.Println(" 3: ",DirEntries[3].Name());
//                                                       fmt.Println(" 2: ",DirEntries[2].Name());
//                                                       fmt.Println();

  for i, finfo := range DirEntries {
    if finfo != nil {
      fmt.Printf("%d: %s, %v, %d bytes, %d nanosecs \n",i,finfo.Name(),finfo.Mode(),finfo.Size(),finfo.ModTime().UnixNano());
    }
  }
  fmt.Println();
  fmt.Println(" Number of filenames is",len(FileNames));
  for i, fname :=  range FileNames {
    fmt.Printf("%d: %s ",i,fname);
  }
  fmt.Println();
  for i, fname :=  range FileNames {
    if fname != "" {
      fmt.Printf("%d: %s ",i,fname);
    }
  }
  fmt.Println();
  fmt.Println(FileNames);
  fmt.Println(" Time now: ",time.Now());
  fmt.Println();
  fmt.Println();

  shell := os.Getenv("SHELL");
  home := os.Getenv("HOME");
  user := os.Getenv("USER");
  gopath := os.Getenv("GOPATH");
  workingdir := os.Getenv("PWD");
  fmt.Println("shell: ",shell,", home: ",home,", user: ",user,", gopath: ",gopath,", Working Dir: ",workingdir);
  fmt.Println();
  println();
}

/*
package path
func Match

func Match(pattern, name string) (matched bool, err error)

Match reports whether name matches the shell file name pattern.  The pattern syntax is:

pattern:
	{ term }
term:
	'*'         matches any sequence of non-/ characters
	'?'         matches any single non-/ character
	'[' [ '^' ] { character-range } ']'
	            character class (must be non-empty)
	c           matches character c (c != '*', '?', '\\', '[')
	'\\' c      matches character c

character-range:
	c           matches character c (c != '\\', '-', ']')
	'\\' c      matches character c
	lo '-' hi   matches character c for lo <= c <= hi

Match requires pattern to match all of name, not just a substring.  The only possible returned error is ErrBadPattern, when pattern is malformed. 


package os
type FileInfo

type FileInfo interface {
        Name() string       // base name of the file
        Size() int64        // length in bytes for regular files; system-dependent for others
        Mode() FileMode     // file mode bits
        ModTime() time.Time // modification time
        IsDir() bool        // abbreviation for Mode().IsDir()
        Sys() interface{}   // underlying data source (can return nil)
}

A FileInfo describes a file and is returned by Stat and Lstat.

func Lstat

func Lstat(name string) (FileInfo, error)

Lstat returns a FileInfo describing the named file.  If the file is a symbolic link, the returned FileInfo describes the symbolic link.  Lstat makes no attempt to follow the link.  
If there is an error, it will be of type *PathError.

func Stat

func Stat(name string) (FileInfo, error)

Stat returns a FileInfo describing the named file.  If there is an error, it will be of type *PathError. 


The insight I had with my append troubles that the 1 slice entries were empty, is that when I used append, it would do just that to the end of the slice, and ignore the empty slices.
I needed to make the slice as empty for this to work.  So I am directly assigning the DirEntries slice, and appending the FileNames slice, to make sure that these both are doing what I want.
This code is now doing exactly what I want.  I guess there is no substitute for playing with myself.  Wait, that didn't come out right.  Or did it.

*/
