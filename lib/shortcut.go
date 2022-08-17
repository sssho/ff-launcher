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

func NewShortcut(dir string, path string, tpath string, modtime time.Time, org Origin) (s Shortcut, err error) {
	s.Path = path
	s.TPath = tpath
	s.ModTime = modtime
	s.Org = org
	_, err = os.Stat(tpath)
	if err != nil {
		return s, err
	}
	isdir, err := isDir(tpath)
	if err != nil {
		isdir = false
	}
	var parent string
	if !isdir {
		parent = filepath.Dir(tpath)
	}
	s.Args = "" // TODO
	s.IsDir = isdir
	s.Parent = parent
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
		tpath, err := ResolveShortcut(f)
		if err != nil {
			continue
		}
		if tpath == "" {
			continue
		}
		finfo, err := dentry.Info()
		if err != nil {
			continue
		}
		shortcut, err := NewShortcut(dir, path, tpath, finfo.ModTime(), origin)
		if err != nil {
			continue
		}
		shortcuts = append(shortcuts, shortcut)
	}
	return shortcuts, nil
}

func isDir(tpath string) (bool, error) {
	info, err := os.Stat(tpath)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return true, nil
	} else {
		return false, nil
	}
}

func GetShortcutTexts(shortcuts []Shortcut) []string {
	texts := make([]string, 0, len(shortcuts))
	checkDuplicate := make(map[string]bool)
	for _, s := range shortcuts {
		key := strings.TrimSpace(s.TPath + s.Args)
		if checkDuplicate[key] {
			continue
		}
		// Add parent
		if !s.IsDir {
			if !checkDuplicate[s.Parent] {
				texts = append(texts, fmt.Sprintf("%s %s", folderPrefix, s.Parent))
				checkDuplicate[s.Parent] = true
			}
		}
		texts = append(texts, s.Text())
		checkDuplicate[key] = true
	}

	return texts
}
