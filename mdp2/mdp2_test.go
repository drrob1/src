package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const inputFile = "./testdata/test1.md"

//const resultFile = "test1.md.html"
const goldenFile = "./testdata/test1.md.html"

func TestParseContent(t *testing.T) {
	input, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatal(err)
	}

	result := parseContent(input)
	result = append(result, '\n') // this makes the byte by byte comparison succeed.

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(expected, result) {
		t.Logf(" Parse: len of expected is %d; golden: \n%s\n", len(expected), expected)
		t.Logf(" Parse: len of result is %d; result:\n%s\n", len(result), result)
		t.Error("Result content does not match golden file")
	}
}

func TestRun(t *testing.T) {
	var mockStdOut bytes.Buffer
	if err := run(inputFile, &mockStdOut); err != nil {
		t.Fatal(err)
	}

	resultFile := strings.TrimSpace(mockStdOut.String())
	result, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatal(err)
	}
	result = append(result, '\n') // this makes the byte by byte comparison succeed.

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(expected, result) {
		t.Logf(" Run: len of expected is %d; golden: \n%s\n", len(expected), expected)
		t.Logf(" Run: len of result is %d; result:\n%s\n", len(result), result)
		t.Error("Result content does not match golden file")
	}

	_ = os.Remove(resultFile)
}
