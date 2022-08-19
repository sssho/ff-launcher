package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Shortcut struct {
	Path    string
	TPath   string
	Args    string
	IsDir   bool
	Parent  string
	Org     Origin
	ModTime time.Time
}

func NewShortcut(path string, tpath string, args string, isdir bool, modtime time.Time, org Origin) (s Shortcut, err error) {
	_, err = os.Stat(tpath)
	if err != nil {
		return s, err
	}
	s.Path = path
	s.TPath = tpath
	s.Args = args
	s.IsDir = isdir
	s.Parent = ""
	s.Org = org
	s.ModTime = modtime
	return s, nil
}

func (s Shortcut) Text() (text string) {
	if s.IsDir {
		text = fmt.Sprintf("%s %s", folderPrefix, s.TPath)
	} else {
		text = fmt.Sprintf("%s %s %s", filePrefix, s.TPath, s.Args)
	}
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
		tpath, isdir, args, err := ResolveShortcut(f)
		if err != nil {
			continue
		}
		finfo, err := dentry.Info()
		if err != nil {
			continue
		}
		shortcut, err := NewShortcut(path, tpath, args, isdir, finfo.ModTime(), origin)
		if err != nil {
			continue
		}
		shortcuts = append(shortcuts, shortcut)
	}
	return shortcuts, nil
}
