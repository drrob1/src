package timlibg

import (
	"fmt"
	"testing"
	"time"
)

func TestJulian(t *testing.T) { // these all pass
	var tests = []struct {
		m, d, y int
		julian  int
	}{
		{1, 1, 1970, 719163},
		{1, 1, 70, 719163},
		{11, 29, 25, 739584},
		{11, 29, 2025, 739584},
		{1, 1, 2025, 739252},
		{1, 1, 25, 739252},
		{12, 1, 2025, 739586},
		{12, 1, 25, 739586},
	}
	fmt.Printf("Testing Julian() with %d tests\n", len(tests))

	for _, test := range tests {
		julian := JULIAN(test.m, test.d, test.y)
		if julian != test.julian {
			t.Errorf("Julian(%d, %d, %d) = %d, want %d", test.m, test.d, test.y, julian, test.julian)
		}
	}
}

func TestSecToHMS(t *testing.T) { // these tests were mostly written by AI.
	var tests = []struct {
		seconds int
		Hours   int
		Minutes int
		Seconds int
	}{
		{3661, 1, 1, 1},
		{61, 0, 1, 1},
		{121, 0, 2, 1},
		{3601, 1, 0, 1},
	}
	fmt.Printf("Testing SecToHMS() with %d tests\n", len(tests))

	for _, test := range tests {
		Hours, Minutes, Seconds := SecToHMS(test.seconds)
		if Hours != test.Hours || Minutes != test.Minutes || Seconds != test.Seconds {
			t.Errorf("SecToHMS(%d) = %d:%d:%d, want %d:%d:%d", test.seconds, Hours, Minutes, Seconds, test.Hours, test.Minutes, test.Seconds)
		}
	}
}

func TestGetDateTime(t *testing.T) {
	now := GetDateTime()
	juldate := Julian(now.Month, now.Day, now.Year)
	lastYrJuldate := Julian(12, 31, now.Year-1)
	myYearDay := juldate - lastYrJuldate
	if myYearDay != now.YearDayNum {
		t.Errorf("myYearDay = %d, want %d", myYearDay, now.YearDayNum)
	}
	if now.WeekDayNum != juldate%7 {
		t.Errorf("juldate mod 7 = %d, want %d", juldate%7, now.WeekDayNum)
	}
	july4, _ := time.Parse(time.DateOnly, "2026-07-04")
	july4DayNum := july4.YearDay()
	JuldateJuly4 := Julian(7, 4, 2026) - lastYrJuldate
	if july4DayNum != JuldateJuly4 {
		t.Errorf("july4DanNum = %d, JuldateJuly4 = %d", july4DayNum, JuldateJuly4)
	}
}
