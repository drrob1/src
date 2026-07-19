package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/tealeg/xlsx/v3"
)

var (
	wordRE   = regexp.MustCompile(`[A-Za-z][A-Za-z'-]*`)
	doctorRE = regexp.MustCompile(`(?i)\bdr\.?\s+([A-Za-z][A-Za-z'-]*)`)
	remoteRE = regexp.MustCompile(`(?i)(?:dr\.?\s+)?([A-Za-z][A-Za-z'-]*)\s*\(\*R\)`)
)

type candidate struct {
	path    string
	modTime time.Time
}

type position struct {
	row int
	col int
}

type finding struct {
	kind        string
	day         string
	name        string
	source      position
	conflict    position
	conflictRow string
}

func main() {
	if err := run(os.Stdin, os.Stdout, time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(input io.Reader, output io.Writer, now time.Time) error {
	files, err := discover(".", now, 30*24*time.Hour)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		files, err = discover(".", now, 90*24*time.Hour)
		if err != nil {
			return err
		}
	}
	if len(files) == 0 {
		fmt.Fprintln(output, "No .xlsx files containing \"week\" were modified in the last 90 days.")
		return nil
	}

	chosen, err := chooseFile(input, output, files)
	if err != nil {
		return err
	}
	findings, err := analyzeFile(chosen)
	if err != nil {
		return err
	}
	if len(findings) == 0 {
		fmt.Fprintln(output, "No schedule errors found.")
		return nil
	}

	for _, f := range findings {
		fmt.Fprintf(output, "%s | %s | %s | source %s | conflict %s (%s)\n",
			f.kind, f.day, f.name, cellReference(f.source),
			cellReference(f.conflict), f.conflictRow)
	}
	return nil
}

func discover(root string, now time.Time, age time.Duration) ([]candidate, error) {
	var files []candidate
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		name := strings.ToLower(entry.Name())
		if filepath.Ext(name) != ".xlsx" || !strings.Contains(name, "week") {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.ModTime().Before(now.Add(-age)) && !info.ModTime().After(now) {
			files = append(files, candidate{path: path, modTime: info.ModTime()})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan for workbooks: %w", err)
	}
	slices.SortFunc(files, func(a, b candidate) int {
		return b.modTime.Compare(a.modTime)
	})
	return files, nil
}

func chooseFile(input io.Reader, output io.Writer, files []candidate) (string, error) {
	for i, file := range files {
		fmt.Fprintf(output, "%d. %s (%s)\n", i+1, file.path, file.modTime.Format("2006-01-02 15:04"))
	}

	scanner := bufio.NewScanner(input)
	for {
		fmt.Fprint(output, "Select a file number: ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", fmt.Errorf("read selection: %w", err)
			}
			return "", errors.New("no file selection provided")
		}
		selection, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err == nil && selection >= 1 && selection <= len(files) {
			return files[selection-1].path, nil
		}
		fmt.Fprintln(output, "Please enter a number from the list.")
	}
}

func analyzeFile(path string) ([]finding, error) {
	book, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	if len(book.Sheets) == 0 {
		return nil, errors.New("workbook has no worksheets")
	}

	sheet := book.Sheets[0]
	cells := make([][]string, sheet.MaxRow)
	for row := range sheet.MaxRow {
		cells[row] = make([]string, sheet.MaxCol)
		for col := range sheet.MaxCol {
			cell, err := sheet.Cell(row, col)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", cellReference(position{row, col}), err)
			}
			cells[row][col] = cell.String()
		}
	}
	return analyze(cells)
}

func analyze(cells [][]string) ([]finding, error) {
	office, err := findLabel(cells, "MDs Out of Office")
	if err != nil {
		return nil, err
	}
	late, err := findLabel(cells, "Late MD")
	if err != nil {
		return nil, err
	}
	fluoroFH, err := findLabel(cells, "FLUORO FH")
	if err != nil {
		return nil, err
	}
	fluoroJH, err := findLabel(cells, "FLUORO JH")
	if err != nil {
		return nil, err
	}

	var findings []finding
	for col := office.col + 1; col < len(cells[office.row]); col++ {
		day := columnDay(cells, col)
		officeNames := words(cellAt(cells, office.row, col))
		for _, name := range officeNames {
			for row := 0; row < office.row; row++ {
				for _, other := range words(cellAt(cells, row, col)) {
					if sameName(name, other) {
						findings = append(findings, newFinding("Error 1", day, name,
							position{office.row, col}, position{row, col}, rowLabel(cells, row)))
					}
				}
			}
		}

		fluoroNames := append(extractDoctorNames(cellAt(cells, fluoroFH.row, col)),
			extractDoctorNames(cellAt(cells, fluoroJH.row, col))...)
		fluoroRows := append(repeatPosition(fluoroFH.row, col, len(extractDoctorNames(cellAt(cells, fluoroFH.row, col)))),
			repeatPosition(fluoroJH.row, col, len(extractDoctorNames(cellAt(cells, fluoroJH.row, col))))...)

		for row := 0; row < office.row; row++ {
			if row == late.row {
				continue
			}
			for _, remote := range extractRemoteNames(cellAt(cells, row, col)) {
				for i, fluoroscopy := range fluoroNames {
					if sameName(remote, fluoroscopy) {
						findings = append(findings, newFinding("Error 2", day, remote,
							position{row, col}, fluoroRows[i], rowLabel(cells, fluoroRows[i].row)))
					}
				}
			}
		}

		for _, lateName := range extractLateNames(cellAt(cells, late.row, col)) {
			for i, fluoroscopy := range fluoroNames {
				if sameName(lateName, fluoroscopy) {
					findings = append(findings, newFinding("Error 3", day, lateName,
						position{late.row, col}, fluoroRows[i], rowLabel(cells, fluoroRows[i].row)))
				}
			}
		}
	}
	return findings, nil
}

func newFinding(kind, day, name string, source, conflict position, conflictRow string) finding {
	return finding{kind: kind, day: day, name: name, source: source, conflict: conflict, conflictRow: conflictRow}
}

func findLabel(cells [][]string, label string) (position, error) {
	target := normalizeLabel(label)
	for row, values := range cells {
		for col, value := range values {
			if strings.Contains(normalizeLabel(value), target) {
				return position{row, col}, nil
			}
		}
	}

	best := position{-1, -1}
	bestDistance := 3
	for row, values := range cells {
		for col, value := range values {
			distance := editDistance(normalizeLabel(value), target)
			if distance < bestDistance {
				best = position{row, col}
				bestDistance = distance
			}
		}
	}
	if best.row >= 0 {
		return best, nil
	}
	return position{}, fmt.Errorf("could not locate row labeled %q", label)
}

func normalizeLabel(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, value)
}

func extractDoctorNames(value string) []string {
	matches := doctorRE.FindAllStringSubmatch(value, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		names = append(names, match[1])
	}
	return names
}

func extractRemoteNames(value string) []string {
	matches := remoteRE.FindAllStringSubmatch(value, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		names = append(names, match[1])
	}
	return names
}

func extractLateNames(value string) []string {
	names := extractDoctorNames(value)
	for _, remote := range extractRemoteNames(value) {
		if !containsSameName(names, remote) {
			names = append(names, remote)
		}
	}
	return names
}

func containsSameName(names []string, wanted string) bool {
	for _, name := range names {
		if sameName(name, wanted) {
			return true
		}
	}
	return false
}

func words(value string) []string {
	return wordRE.FindAllString(value, -1)
}

func sameName(a, b string) bool {
	a = strings.ToLower(strings.Trim(a, "-'"))
	b = strings.ToLower(strings.Trim(b, "-'"))
	return a != "" && b != "" && (editDistance(a, b) <= 1 || oneAdjacentTransposition(a, b))
}

func oneAdjacentTransposition(a, b string) bool {
	left, right := []rune(a), []rune(b)
	if len(left) != len(right) {
		return false
	}
	first := -1
	for i := range left {
		if left[i] == right[i] {
			continue
		}
		if first < 0 {
			first = i
			continue
		}
		return i == first+1 && left[first] == right[i] && left[i] == right[first] &&
			string(left[i+1:]) == string(right[i+1:])
	}
	return false
}

func editDistance(a, b string) int {
	left, right := []rune(a), []rune(b)
	previous := make([]int, len(right)+1)
	for i := range previous {
		previous[i] = i
	}
	for i, l := range left {
		current := make([]int, len(right)+1)
		current[0] = i + 1
		for j, r := range right {
			cost := 0
			if l != r {
				cost = 1
			}
			current[j+1] = min(current[j]+1, previous[j+1]+1, previous[j]+cost)
		}
		previous = current
	}
	return previous[len(right)]
}

func columnDay(cells [][]string, col int) string {
	weekday := cellAt(cells, 1, col)
	date := cellAt(cells, 2, col)
	if weekday == "" {
		return date
	}
	if date == "" {
		return weekday
	}
	return weekday + " " + date
}

func cellAt(cells [][]string, row, col int) string {
	if row < 0 || row >= len(cells) || col < 0 || col >= len(cells[row]) {
		return ""
	}
	return cells[row][col]
}

func rowLabel(cells [][]string, row int) string {
	for _, value := range cells[row] {
		if strings.TrimSpace(value) != "" {
			return strings.Join(strings.Fields(value), " ")
		}
	}
	return "unlabeled row"
}

func repeatPosition(row, col, count int) []position {
	positions := make([]position, count)
	for i := range positions {
		positions[i] = position{row, col}
	}
	return positions
}

func cellReference(p position) string {
	col := p.col + 1
	var letters string
	for col > 0 {
		col--
		letters = string(rune('A'+col%26)) + letters
		col /= 26
	}
	return letters + strconv.Itoa(p.row+1)
}
