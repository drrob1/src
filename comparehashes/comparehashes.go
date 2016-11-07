package main;

import (
"os"
"bufio"
"fmt"
"runtime"
"strings"
"encoding/hex"
"crypto/sha512"
"crypto/sha256"
"crypto/sha1"
"crypto/md5"
"io"
"path/filepath"
"hash"
"getcommandline"
"tokenize"
)
/*
  REVISION HISTORY
  ----------------
   6 Apr 13 -- First modified version of module.  I will use VLI to compare all digits of the hashes.
  23 Apr 13 -- Fixed problem of a single line in the hashes file, that does not contain an EOL character, causes
                an immediate return without processing of the characters just read in.
  24 Apr 13 -- Added output of which file either matches or does not match.
  19 Sep 16 -- Finished conversion to Go, that was started 13 Sep 16.  Added the removed of '*' which is part of a std linux formated hash file.  And I forgot that
                 the routine allowed either order in the file.  If the token has a '.' I assume it is a filename, else it is a hash value.
  21 Sep 16 -- Fixed the case issue in tokenize.GetToken.  Edited code here to correspond to this fix.
*/


//* ************************* MAIN ***************************************************************
func main() {


  const K = 1024;
  const M = 1024*1024;

  const (
         md5hash = iota
         sha1hash
         sha256hash
         sha384hash
         sha512hash
         HashType
        );

  const ReadBufferSize = M;
//  const ReadBufferSize = 80 * M;

  var HashName = [...]string{"md5","sha1","sha256","sha384","sha512"};
  var inbuf string;
  var WhichHash int;
  var readErr error;
  var TargetFilename,HashValueReadFromFile,HashValueComputedStr string;
  var hasher hash.Hash;
  var FileSize int64;


  if len(os.Args) <= 1 {
    fmt.Println(" Usage: comparehashes <hashFileName.ext> where .ext = [.md5|.sha1|.sha256|.sha384|.sha512]");
    os.Exit(0);
  }
  inbuf = getcommandline.GetCommandLineString();

  extension := filepath.Ext(inbuf);
  extension = strings.ToLower(extension);
  switch extension {
  case ".md5" :
    WhichHash = md5hash;
  case ".sha1":
    WhichHash = sha1hash;
  case ".sha256":
    WhichHash = sha256hash;
  case ".sha384":
    WhichHash = sha384hash;
  case ".sha512":
    WhichHash = sha512hash;
  default:
    fmt.Println();
    fmt.Println();
    fmt.Println(" Not a recognized hash extension.  Will assume sha1.");
    WhichHash = sha1hash;
  } // switch case on extension for HashType 

  fmt.Println();
  fmt.Println(" GOOS =",runtime.GOOS,".  ARCH=",runtime.GOARCH);
  fmt.Println();
  fmt.Println();
  fmt.Println();

  fmt.Print(" Testing determining hash type by file extension.");
  fmt.Println("  HashType = md5, sha1, sha256, sha384, sha512.  WhichHash = ",HashName[WhichHash]);
  fmt.Println();


// Read and parse the file with the hashes.

  HashesFile,err := os.Open(inbuf);
  check(err,"Cannot open HashesFile.  Does it exist?  ");
  defer HashesFile.Close();

  scanner := bufio.NewScanner(HashesFile);
  scanner.Split(bufio.ScanLines);    // I believe this is the default.  I may experiment to see if I need this line for my code to work, AFTER I debug it as it is.

  for { /* to read multiple lines */
    FileSize = 0;
    readSuccess := scanner.Scan();
    if !readSuccess {break}    // end main reading loop

    inputline := scanner.Text();
    if readErr = scanner.Err(); readErr != nil {
      if readErr == io.EOF { break }  // reached EOF condition, so there are no more lines to read.
      fmt.Fprintln(os.Stderr, "Unknown error while reading from the HashesFile :", readErr);
      os.Exit(1);
    }

    if strings.HasPrefix(inputline,";") || strings.HasPrefix(inputline,"#") || (len(inputline) <= 10) { continue } /* allow comments and essentially blank lines */

    tokenize.INITKN(inputline);

    FirstToken,EOL := tokenize.GetTokenString(false);

    if EOL  {
      fmt.Errorf(" Error while getting 1st token in the hashing file.  Skipping");
      continue;
    }

    if strings.ContainsRune(FirstToken.Str,'.') { /* have filename first on line */
      TargetFilename = FirstToken.Str;
      SecondToken,EOL := tokenize.GetTokenString(false);  // Get hash string from the line in the file
      if EOL {
        fmt.Errorf(" Got EOL while getting HashValue (2nd) token in the hashing file.  Skipping \n");
        continue;
      } /* if EOL */
      HashValueReadFromFile = SecondToken.Str;

    }else{  /* have hash first on line */
      HashValueReadFromFile = FirstToken.Str;
      SecondToken,EOL := tokenize.GetTokenString(false);   // Get name of file on which to compute the hash
      if EOL {
        fmt.Errorf(" Error while gatting TargetFilename token in the hashing file.  Skipping \n");
        continue;
      } /* if EOL */

      if strings.ContainsRune(SecondToken.Str,'*') {  // If it contains a *, it will be the first position.
        SecondToken.Str = SecondToken.Str[1:];
        if strings.ContainsRune(SecondToken.Str,'*') { // this should not happen
          fmt.Println(" Filename token still contains a * character.  Str:",SecondToken.Str,"  Skipping.");
          continue;
        }
      }
      TargetFilename = (SecondToken.Str);
    } /* if have filename first or hash value first */

/*
  now to compute the hash, compare them, and output results
*/
    /* Create Hash Section */
    TargetFile,err := os.Open(TargetFilename);
    check(err," Error opening TargetFilename.");
    defer TargetFile.Close();

    switch WhichHash {     // Initialing case switch on WhichHash
    case md5hash :
           hasher = md5.New();
    case sha1hash :
           hasher = sha1.New();
    case sha256hash :
           hasher = sha256.New();
    case sha384hash :
           hasher = sha512.New384();
    case sha512hash :
           hasher = sha512.New();
    default:
           hasher = sha256.New();
    } /* initializing switch case on WhichHash */

/*    This loop works, but there is a much shorter way.  I got this after asking for help on the mailing list.
    FileReadBuffer := make([]byte,ReadBufferSize);
    for {   // Repeat Until eof loop.
      n,err := TargetFile.Read(FileReadBuffer);
      if n == 0 || err == io.EOF { break }
      check(err," Unexpected error while reading the target file on which to compute the hash,");
      hasher.Write(FileReadBuffer[:n]);
      FileSize += int64(n);
    } // Repeat Until TargetFile.eof loop;
*/

    FileSize,readErr = io.Copy(hasher,TargetFile);
    HashValueComputedStr = hex.EncodeToString(hasher.Sum(nil));

// I got the idea to use the different base64 versions and my own hex converter code, just to see.
// And I can also use sfprintf with the %x verb.  base64 versions and not useful as they use a larger
// character set than hex.  I deleted all references to the base64 versions.  And the hex encoded and
// sprintf using %x were the same, so I removed the sprintf code.
//    HashValueComputedSprintf := fmt.Sprintf("%x",hasher.Sum(nil));

    fmt.Println(" Filename  = ",TargetFilename,", FileSize = ",FileSize,", ",HashName[WhichHash]," computed hash string, followed by hash string in the file are : ");
    fmt.Println("       Read From File:",HashValueReadFromFile);
    fmt.Println(" Computed hex encoded:",HashValueComputedStr);
//    fmt.Println(" Computed sprintf:",HashValueComputedSprintf);

    if HashValueReadFromFile == HashValueComputedStr {
      fmt.Print(" Matched.");
    }else{
      fmt.Print(" Not matched.");
    } /* if hashes */
    TargetFile.Close();     // Close the handle to allow opening a target from the next line, if there is one.
    fmt.Println();
    fmt.Println();
    fmt.Println();
  }   /* outer LOOP to read multiple lines */

  HashesFile.Close();    // Don't really need this because of the defer statement.
  fmt.Println();
}  // Main for comparehashes.go.

// ------------------------------------------------------- check -------------------------------
func check(e error, msg string) {
  if e != nil {
    fmt.Errorf("%s : ",msg);
    panic(e);
  }
}

