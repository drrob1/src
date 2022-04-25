package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestRun(t *testing.T) {
	var testCases = []struct { // table driven tests
		name   string
		proj   string
		out    string // espected output message
		expErr error
	}{
		{name: "build success", proj: "./testdata/tool/", out: "Go Build: SUCCESS\n", expErr: nil},
		{name: "build fail", proj: "./testdata/toolErr", out: "", expErr: &stepErr{step: "go build"}},
		{name: "test success", proj: "./testdata/tool/", out: "Go Test: SUCCESS\n", expErr: nil},
	}

	for _, tc := range testCases { // range over the table of tests.  Right now there are only 2.
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := run(tc.proj, &out)
			if tc.expErr != nil {
				if err == nil {
					t.Errorf("Expected error: %q.  Got 'nil' instead.", tc.expErr)
					return
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error: %q.  Got %q.", tc.expErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}
			if out.String() != tc.out {
				t.Errorf("Expected output: %q.  Got %q", tc.out, out.String())
			}
		})
	}
}
