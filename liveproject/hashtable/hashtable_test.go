package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) { // this example is in the docs of the testing package, that I was referred to by the golang nuts google group.
	os.Exit(m.Run())
}

func main() {

}
