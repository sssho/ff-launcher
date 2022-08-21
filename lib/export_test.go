package lib

import "time"

func (h *HistItem) SetPath(path string) {
	h.path = path
}

func NewHistItem(path string, isdir bool, time time.Time) *HistItem {
	return &HistItem{path, isdir, time}
}
