package main;

import (
"os"
"fmt"
"runtime"
//"strings"
"encoding/xml"
"io"
"path"
//"path/filepath"
"getcommandline"

)


/*
  REVISION HISTORY
  ----------------
  22 Sep 16 -- I started coding this.  This may take a while.  Added code from Donovan &
Kerninghan book.
*/


func main (){

var tokenstring string;

  if len(os.Args) <= 1 {
    fmt.Println(" Usage: comparehashes <hashFileName.ext> where .ext = [.md5|.sha1|.sha256|.sha384|.sha512]");
    os.Exit(0);
  }
  inbuf := getcommandline.GetCommandLineString();  // inbuf is a string

  filename := path.Clean(inbuf);
  fmt.Println(" The filename to read, parse and display is:",filename);

  fmt.Println();
  fmt.Println(" GOOS =",runtime.GOOS,".  ARCH=",runtime.GOARCH);
  fmt.Println();
  xmlfile,err := os.Open(filename);
  if err != nil {
    fmt.Errorf(" Unable to open %s.  Does it exist?\n",filename);
    os.Exit(1);
  }
  defer xmlfile.Close();

  loopctr := 0;
  xmldecoder := xml.NewDecoder(xmlfile);

  for {  // read file until can't anymore
    loopctr++;
    fmt.Println();
    fmt.Println("LoopCtr:",loopctr);
    xmltoken,err := xmldecoder.Token();
    if err == io.EOF {
      fmt.Println(" err = io.EOF");
      break;
    }else if err != nil {
      fmt.Errorf(" getting a token and err not io.EOF and not nil\n");
      break;
    }else if xmltoken == nil {
      fmt.Println(" token = nil but err was not io.EOF");
      break;
    }

    switch xmltoken.(type) {
    case xml.StartElement:
      fmt.Println(" StartElement as raw bytes:",xmltoken);
      startelement :=  xmltoken.(xml.StartElement).Copy();  // needed the type assertion and method syntax to compile
      fmt.Printf(" startelement token type is %T,\n raw contents %v,\n as string %s\n",startelement,startelement,startelement);
      tokenstring = fmt.Sprintf("%s",startelement);
      fmt.Printf(" converted StartElement to tokenstring using Sprintf.  tokenstring: %s\n",tokenstring);
      startelementspacename := xmltoken.(xml.StartElement).Name.Space;
      startelementlocalname := xmltoken.(xml.StartElement).Name.Local;
      fmt.Printf(" Name space type is %T, value as string is %s\n",startelementspacename,startelementspacename);
      fmt.Printf(" Name local type is %T, value as string is %s\n",startelementlocalname,startelementlocalname);

    case xml.EndElement:
      fmt.Printf(" EndElement token type is %T,  as raw bytes: %v,\n as string %s \n",xmltoken,xmltoken,xmltoken);
//      tokenstring = xmltoken.(xml.EndElement).Copy();  // won't compile, no way, no how.
      tokenstring = fmt.Sprintf("%s",xmltoken);
      fmt.Printf(" converted EndElement to tokenstring using Sprintf.  tokenstring: %s\n",tokenstring);
      endelementspacename := xmltoken.(xml.EndElement).Name.Space;
      endelementlocalname := xmltoken.(xml.EndElement).Name.Local;
      fmt.Printf(" Name space type is %T, value as string is %s\n",endelementspacename,endelementspacename);
      fmt.Printf(" Name local type is %T, value as string is %s\n",endelementlocalname,endelementlocalname);

    case xml.CharData:
      fmt.Println(" CharData as raw bytes:",xmltoken);
      tokenstring = string(xml.CopyToken(xmltoken).(xml.CharData))  // needed the type assertion to compile;
      fmt.Printf(" CharData tokenstring using string verb is %s\n",tokenstring);
// Looking to eliminate the control characters like <lf>, <tab>, ' ', and whatever other garbage is in the file.
// No-go
/*
      skiperr := xmldecoder.Skip();
      if skiperr == nil {
        fmt.Println("  Successful skip")
      }else{
        fmt.Println("Unsuccessful skip.  Error is:",skiperr);
      }
*/

    case xml.Comment:
      fmt.Printf(" Comment as raw bytes: %v,\n as string: %s\n",xmltoken,xmltoken);
      tokenstring = string(xml.CopyToken(xmltoken).(xml.Comment));  // needed the type assertion to compile
      fmt.Printf(" Comment tokenstring using string verb is %s\n",tokenstring);

    case xml.Directive:
      fmt.Printf(" Directive as raw bytes: %v\n as string: %s\n",xmltoken,xmltoken);
      tokenstring = string(xmltoken.(xml.Directive).Copy());
      fmt.Printf(" Directive tokenstring using string verb is %s\n",tokenstring);

    }
    fmt.Println();
  } // end for
  fmt.Println();
  fmt.Println();
  fmt.Println(" Finished.  The input file closed by defer.");
  fmt.Println();
}
