package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Shortcut struct {
	TPath   string
	IsDir   bool
	ModTime time.Time
}

func NewShortcut(tpath string, isdir bool, modtime time.Time) (s Shortcut, err error) {
	_, err = os.Stat(tpath)
	if err != nil {
		return s, err
	}
	s.TPath = tpath
	s.IsDir = isdir
	s.ModTime = modtime
	return s, nil
}

func (s Shortcut) Text() (text string) {
	text = fmt.Sprintf("%s %s", folderPrefix, s.TPath)
	text = strings.TrimSpace(text)
	return
}

func NewShortcuts(dir string, origin Origin) ([]Shortcut, error) {
	dentries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	shortcuts := make([]Shortcut, 0, len(dentries))
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
		shortcut, err := NewShortcut(tpath, isdir, finfo.ModTime())
		if err != nil {
			continue
		}
		shortcuts = append(shortcuts, shortcut)
	}
	return shortcuts, nil
}
