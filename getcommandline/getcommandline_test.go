package getcommandline;

import "testing";
import "fmt";

func Testgetcommandline(t *testing.T) {
  TheString := GetCommandLineString()
  fmt.Println(" Input commandline is : ",TheString);
}
