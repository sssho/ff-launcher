package lib

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	tpath, args, err := GetShortcutInfo(s.Path, w)
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

func worker(input chan fs.FileInfo, output chan Shortcut, dir string, org Origin, w *WscriptShell, wg *sync.WaitGroup) {
	for finfo := range input {
		shortcut, err := NewShortcut(dir, finfo, org)
		if err != nil {
			wg.Done()
			continue
		}
		err = shortcut.Resolve(w)
		if err != nil {
			wg.Done()
			continue
		}
		output <- *shortcut
		wg.Done()
	}
}

func NewShortcutList(dir string, origin Origin) ([]Shortcut, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	input := make(chan fs.FileInfo, len(files))
	output := make(chan Shortcut)

	for _, f := range files {
		input <- f
	}

	shortcuts := make([]Shortcut, 0, len(files))

	ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED|ole.COINIT_DISABLE_OLE1DDE|ole.COINIT_SPEED_OVER_MEMORY)
	defer ole.CoUninitialize()
	w, err := NewWscriptShell()
	if err != nil {
		return nil, err
	}
	defer w.Release()

	var wg sync.WaitGroup
	wg.Add(len(files))
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(input, output, dir, origin, w, &wg)
	}
	close(input)
	go func() {
		for v := range output {
			shortcuts = append(shortcuts, v)
		}
		close(output)
	}()
	wg.Wait()
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
