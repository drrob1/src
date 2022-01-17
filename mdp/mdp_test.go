package main

import (
	"bytes"
	"os"
	"testing"
)

const inputFile = "./testdata/test1.md"
const resultFile = "test1.md.html"
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
	if err := run(inputFile); err != nil {
		t.Fatal(err)
	}

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
