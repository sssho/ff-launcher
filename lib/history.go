package lib

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"time"
)

type History []HistItem

func (h History) Find(target HistItem) (i int, found bool) {
	// h must be sorted by path
	i, found = sort.Find(len(h), func(i int) int {
		return strings.Compare(target.path, h[i].path)
	})
	if found {
		return i, found
	} else {
		return i, false
	}
}

func (h *History) Insert(target HistItem, i int) {
	// TODO: Improve if performance is not good
	if i >= len(*h) {
		(*h) = append((*h), target)
	} else {
		(*h) = append((*h)[:i+1], (*h)[i:]...)
	}
	(*h)[i] = target
}

func (h *History) Merge(target HistItem) {
	// h must be sorted by path
	if i, found := (*h).Find(target); !found {
		h.Insert(target, i)
	}
	// TODO: if found, compare time and update it if target is newwer
}

func (h *History) SortByTime() {
	sort.Slice(*h, func(i, j int) bool {
		return (*h)[i].lastAccess.After((*h)[j].lastAccess)
	})
}

func (h *History) SortByPath() {
	sort.Slice(*h, func(i, j int) bool {
		return (*h)[i].path < (*h)[j].path
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
	err = e.Encode(h)
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
	err = d.Decode(h)
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
