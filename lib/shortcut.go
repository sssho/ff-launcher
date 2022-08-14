package lib

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
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

func NewShortcut(dir string, finfo fs.FileInfo, org Origin) (s *Shortcut, err error) {
	var shortcut Shortcut
	path := filepath.Join(dir, finfo.Name())
	shortcut.Path = path
	shortcut.ModTime = finfo.ModTime()
	shortcut.Org = org

	return &shortcut, nil
}

func (s *Shortcut) Resolve(w *WscriptShell) error {
	tpath, args, err := ResolveShortcut(s.Path, w)
	if err != nil {
		return err
	}
	if tpath == "" {
		return fmt.Errorf("tpath is nil")
	}
	_, err = os.Stat(tpath)
	if err != nil {
		return err
	}
	var isdir bool
	var parent string
	isdir, err = isDir(tpath)
	if err != nil {
		isdir = false
	}
	if !isdir {
		parent = filepath.Dir(tpath)
	}
	s.TPath = tpath
	s.Args = args
	s.IsDir = isdir
	s.Parent = parent

	return nil
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

func worker(input chan fs.FileInfo, output chan Shortcut, dir string, org Origin, w *WscriptShell) {
	for finfo := range input {
		shortcut, err := NewShortcut(dir, finfo, org)
		if err != nil {
			output <- *shortcut
			continue
		}
		_ = shortcut.Resolve(w)
		output <- *shortcut
	}
}

func NewShortcutList(dir string, origin Origin) ([]Shortcut, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	workNum := len(files)
	input := make(chan fs.FileInfo, workNum)
	output := make(chan Shortcut)

	for _, f := range files {
		input <- f
	}

	shortcuts := make([]Shortcut, 0, workNum)

	ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED|ole.COINIT_DISABLE_OLE1DDE|ole.COINIT_SPEED_OVER_MEMORY)
	defer ole.CoUninitialize()
	w, err := NewWscriptShell()
	if err != nil {
		return nil, err
	}
	defer w.Release()
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(input, output, dir, origin, w)
	}
	// close(input)
	end := make(chan bool)
	go func() {
		var done int
		for v := range output {
			done++
			if v.TPath == "" {
				if done == workNum {
					break
				}
				continue
			}
			shortcuts = append(shortcuts, v)
			if done == workNum {
				break
			}
		}
		end <- true
		// close(output)
	}()
	<-end
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
