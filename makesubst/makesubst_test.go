package makesubst;
/*
I'm going to take notes about this testing stuff.  The Println and Log output was only seen when I used
  go test makesubst -v

from the GOPATH ~/gocode dir.  Without the -v switch, I saw no output other than the OK or pass output.

*/
import (
"testing"
"fmt"
)

func TestMakeSubst(t *testing.T) {
  s1 := MakeSubst("=");
  fmt.Println("fmt.Println -- s1 should be +, and it is ",s1);
  t.Log(" s1 should be + and it is ",s1);
  if s1 != "+" {
    t.Fail();
  }
  s2 := MakeSubst(";");
  t.Log(" s2 should be *, and it is ",s2);
  if s2 != "*" {
    t.Fail();
  }
  s1 = MakeSubst("");
  if s1 != "" {
    t.Fail();
  }
  s3 := MakeSubst(" this is a =1234 string; that I am testing.");
  fmt.Println("fmt.Println -- s3 is ",s3);
  t.Log("t.Log -- s3 is ",s3);
  if s3 != "this is a +1234 string* that I am testing." {
    t.Fail();
  }
}
