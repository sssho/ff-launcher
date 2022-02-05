package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
)

type ShortcutInfo struct {
	Path    string
	TPath   string
	Args    string
	IsDir   bool
	Parent  string
	Org     Origin
	ModTime time.Time
}

func (s ShortcutInfo) Text() (text string) {
	if s.IsDir {
		text = fmt.Sprintf("%s %s", folderPrefix, s.TPath)
	} else {
		text = fmt.Sprintf("%s %s %s", filePrefix, s.TPath, s.Args)
	}
	text = strings.TrimSpace(text)
	return
}

func NewShortcutInfoList(dir string, origin Origin) ([]ShortcutInfo, error) {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	defer ole.CoUninitialize()

	w, err := NewWscriptShell()
	if err != nil {
		return nil, err
	}
	defer w.Release()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	shortcuts := make([]ShortcutInfo, 0, len(files))
	var isdir bool
	var parent string
	for _, file := range files {
		tpath, args, err := GetShortcutInfo(w, filepath.Join(dir, file.Name()))
		if err != nil {
			continue
		}
		if tpath == "" {
			continue
		}
		_, err = os.Stat(tpath)
		if err != nil {
			continue
		}
		isdir, err = isDir(tpath)
		if err != nil {
			isdir = false
		}
		if !isdir {
			parent = filepath.Dir(tpath)
		}
		shortcuts = append(shortcuts, ShortcutInfo{filepath.Join(dir, file.Name()), tpath, args, isdir, parent, origin, file.ModTime()})
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

func GetShortcutTexts(shortcuts []ShortcutInfo) []string {
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
