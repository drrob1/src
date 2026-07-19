package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiscoverFiltersRecursesAndSorts(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)
	createFileAt(t, filepath.Join(root, "older_WEEK.xlsx"), now.Add(-20*24*time.Hour))
	createFileAt(t, filepath.Join(nested, "newer_week.XLSX"), now.Add(-2*24*time.Hour))
	createFileAt(t, filepath.Join(root, "not-a-schedule.xlsx"), now.Add(-time.Hour))
	createFileAt(t, filepath.Join(root, "too-old-week.xlsx"), now.Add(-31*24*time.Hour))

	files, err := discover(root, now, 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
	if filepath.Base(files[0].path) != "newer_week.XLSX" {
		t.Fatalf("first file = %q, want newest file", files[0].path)
	}
	if filepath.Base(files[1].path) != "older_WEEK.xlsx" {
		t.Fatalf("second file = %q, want older file", files[1].path)
	}
}

func TestSameName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b string
		want bool
	}{
		{"Jancuzk", "janczuk", true},
		{"Lee", "LEE", true},
		{"Smith", "Smyth", true},
		{"Smith", "Solomon", false},
	}
	for _, test := range tests {
		if got := sameName(test.a, test.b); got != test.want {
			t.Errorf("sameName(%q, %q) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func TestRemoteMarkerAppliesToImmediatelyPrecedingName(t *testing.T) {
	t.Parallel()

	got := extractRemoteNames("Dr. Lee (@FH) Dr. Choi (*R) Dr. Smith (@JH)")
	if len(got) != 1 || got[0] != "Choi" {
		t.Fatalf("extractRemoteNames() = %v, want [Choi]", got)
	}
}

func TestAnalyzeFindsAllThreeErrorTypes(t *testing.T) {
	t.Parallel()

	cells := [][]string{
		{"ASSIGNMENTS", "Dr. Janczuk"},
		{"", "Monday"},
		{"", "July 20, 2026"},
		{"NEURO", "Dr. Lee (*R)"},
		{"FLUORO JH", "Dr. Choi"},
		{"FLUORO FH", "Dr. Lee"},
		{"JH/FH Late MD", "Dr. Choi (*R)"},
		{"MDs Out of Office", "Jancuzk"},
	}

	findings, err := analyze(cells)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 3 {
		t.Fatalf("got %d findings, want 3: %#v", len(findings), findings)
	}
	for i, kind := range []string{"Error 1", "Error 2", "Error 3"} {
		if findings[i].kind != kind {
			t.Errorf("finding %d kind = %q, want %q", i, findings[i].kind, kind)
		}
		if findings[i].day != "Monday July 20, 2026" {
			t.Errorf("finding %d day = %q", i, findings[i].day)
		}
	}
	if cellReference(findings[0].source) != "B8" || cellReference(findings[0].conflict) != "B1" {
		t.Errorf("Error 1 cells = %s -> %s, want B8 -> B1",
			cellReference(findings[0].source), cellReference(findings[0].conflict))
	}
	if cellReference(findings[1].source) != "B4" || cellReference(findings[1].conflict) != "B6" {
		t.Errorf("Error 2 cells = %s -> %s, want B4 -> B6",
			cellReference(findings[1].source), cellReference(findings[1].conflict))
	}
	if cellReference(findings[2].source) != "B7" || cellReference(findings[2].conflict) != "B5" {
		t.Errorf("Error 3 cells = %s -> %s, want B7 -> B5",
			cellReference(findings[2].source), cellReference(findings[2].conflict))
	}
}

func createFileAt(t *testing.T, path string, modTime time.Time) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
}
