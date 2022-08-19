package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

func NewHistItems(dir string, origin Origin) ([]HistItem, error) {
	dentries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	histItems := make([]HistItem, 0, len(dentries))
	for _, dentry := range dentries {
		path := filepath.Join(dir, dentry.Name())
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()
		tpath, isdir, _, err := ResolveShortcut(f)
		if err != nil {
			continue
		}
		finfo, err := dentry.Info()
		if err != nil {
			continue
		}
		_, err = os.Stat(tpath)
		if err != nil {
			continue
		}
		histItems = append(histItems, HistItem{tpath, isdir, finfo.ModTime()})
	}
	return histItems, nil
}
