package lib

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// util
func CreateHistory(s string) *History {
	var hist History
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		if scanner.Err() != nil {
			break
		}
		h := HistItem{}
		h.SetPath(strings.TrimSpace(scanner.Text()))
		hist = append(hist, h)
	}
	return &hist
}

func CreateHistoryS(items []string) *History {
	var hist History
	for _, item := range items {
		hist = append(hist, HistItem{path: item})

	}
	return &hist
}

func HistString(hist *History) string {
	var b strings.Builder
	for i, v := range *hist {
		fmt.Fprintf(&b, "%d [%v]\n", i, v.path)
	}
	return b.String()
}

func PrintHist(hist *History) {
	for i, v := range *hist {
		fmt.Printf("%d [%v]\n", i, v.path)
	}
}

// test
func TestHistory(t *testing.T) {
	data := []string{
		`C:\Users\a\a0.txt`,
		`C:\Users\a\a1.txt`,
		`C:\Users\b\d0.txt`,
	}
	want := CreateHistoryS(data)
	n := 1
	p := data[n]
	got := CreateHistoryS(append(data[:][:n], data[:][n+1:]...))
	got.Merge(HistItem{path: p})
	if !reflect.DeepEqual(got, want) {
		t.Error("-- got")
		t.Errorf("\n%s", HistString(got))
		t.Error("-- want")
		t.Errorf("\n%s", HistString(want))
	}
}
