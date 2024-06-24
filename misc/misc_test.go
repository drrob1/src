package misc

import "testing"

var testMisc = []struct {
	input  string
	output string
}{
	{"1", "1"},
	{"12", "12"},
	{"123", "123"},
	{"1234", "1,234"},
	{"12345", "12,345"},
	{"123456", "123,456"},
	{"1234567", "1,234,567"},
	{"12345678", "12,345,678"},
	{"123456789", "123,456,789"},
	{"1234567890", "1,234,567,890"},
}

func TestMisc(t *testing.T) {
	for _, v := range testMisc {
		o := AddCommasRune(v.input)
		if o != v.output {
			t.Errorf(" input= %s, output= %s, expected %s", v.input, o, v.output)
		}
	}

}
