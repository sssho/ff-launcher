package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type History struct {
	items []HistItem
}

func (h *History) Append(i HistItem) {

}

func (h *History) Sort() {
	sort.Slice(h.items, func(i, j int) bool {
		return h.items[i].lastAccess.After(h.items[j].lastAccess)
	})
}

func (h History) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	err = e.Encode(h.items)
	if err != nil {
		return err
	}
	return nil
}

func (h *History) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	err = d.Decode(&h.items)
	if err != nil {
		return err
	}
	return nil
}

type HistItem struct {
	path       string
	isDir      bool
	lastAccess time.Time
}

func (h HistItem) Text() (text string) {
	text = fmt.Sprintf("%s %s", folderPrefix, h.path)
	text = strings.TrimSpace(text)
	return
}

func contains(h HistItem, hitems []HistItem) bool {
	for _, hitem := range hitems {
		if h.path == hitem.path {
			return true
		}
	}
	return false
}
