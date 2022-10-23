package cmd

import (
	"os"
	"src/cobra/pscan/scan"
	"testing"
)

func setup(t *testing.T, hosts []string, initList bool) (string, func()) {
	// create temp file
	tf, err := os.CreateTemp("", "pscan")
	if err != nil {
		t.Fatal(err)
	}
	tf.Close()

	// init list if asked by initList input param.
	if initList {
		hl := &scan.HostsList{}

		for _, h := range hosts {
			hl.Add(h)
		}

		if err := hl.Save(tf.Name()); err != nil {
			t.Fatal(err)
		}
	}

	// return temp file name and cleanup function
	return tf.Name(), func() {
		os.Remove(tf.Name())
	}
}

func TestHostActions(t *testing.T) {

}
