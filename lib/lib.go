package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	exitOK    = 0
	exitError = 1
)

const (
	folderPrefix = "[folder]"
	filePrefix   = "[file  ]"
)

type WscriptShell struct {
	Shell  *ole.IUnknown
	Wshell *ole.IDispatch
}

func NewWscriptShell() (*WscriptShell, error) {
	shell, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return nil, err
	}
	wshell, err := shell.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		shell.Release()
		return nil, err
	}
	return &WscriptShell{shell, wshell}, nil
}

func (w *WscriptShell) Release() {
	w.Shell.Release()
	w.Wshell.Release()
}

func (w *WscriptShell) GetShortcutInfo(path string) (string, string, error) {
	shortcut, err := oleutil.CallMethod(w.Wshell, "CreateShortcut", path)
	if err != nil {
		return "", "", err
	}
	shortcutDispath := shortcut.ToIDispatch()

	targetPath, err := shortcutDispath.GetProperty("TargetPath")
	if err != nil {
		return "", "", err
	}

	args, err := shortcutDispath.GetProperty("Arguments")
	if err != nil {
		return "", "", err
	}
	return targetPath.ToString(), args.ToString(), nil
}

type ShortcutInfo struct {
	Path   string
	TPath  string
	Args   string
	IsDir  bool
	Parent string
}

func (s ShortcutInfo) Text() (text string) {
	if s.IsDir {
		text = fmt.Sprintf("%s %s", folderPrefix, s.TPath)
	} else {
		text = fmt.Sprintf("%s %s %s", filePrefix, s.TPath, s.Args)
	}
	return
}

func NewShortcutInfoList(dir string) ([]ShortcutInfo, error) {
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
		tpath, args, err := w.GetShortcutInfo(filepath.Join(dir, file.Name()))
		if err != nil {
			continue
		}
		if tpath == "" {
			continue
		}
		isdir, err = isDir(tpath)
		if err != nil {
			isdir = false
		}
		if !isdir {
			parent = filepath.Dir(tpath)
		}
		shortcuts = append(shortcuts, ShortcutInfo{file.Name(), tpath, args, isdir, parent})
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

func GetRecentDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`%s\AppData\Roaming\Microsoft\Windows\Recent`, home), nil
}

type Source [][]string

func RunFF(sources Source) (string, error) {
	ff, err := exec.LookPath("peco")
	if err != nil {
		return "", err
	}

	cmd := exec.Command(ff)
	cmd.Stderr = os.Stderr
	in, _ := cmd.StdinPipe()
	go func() {
		defer in.Close()

		for _, source := range sources {
			for _, s := range source {
				io.WriteString(in, s+"\n")
			}
		}
	}()
	result, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(result)), nil
}

func OpenExplore(path string) error {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	defer ole.CoUninitialize()

	oleShellObject, err := oleutil.CreateObject("Shell.Application")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()

	_, err = oleutil.CallMethod(wshell, "Explore", path)
	if err != nil {
		return err
	}
	return nil
}

func RunDefaultApp(path string) error {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	defer ole.CoUninitialize()

	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()

	_, err = oleutil.CallMethod(wshell, "Run", fmt.Sprintf("\"%s\"", path))
	if err != nil {
		return err
	}

	return nil
}

func RunApp(path string) error {
	if strings.HasPrefix(path, folderPrefix) {
		err := OpenExplore(strings.TrimSpace(strings.Replace(path, folderPrefix, "", -1)))
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(path, filePrefix) {
		err := RunDefaultApp(strings.TrimSpace(strings.Replace(path, filePrefix, "", -1)))
		if err != nil {
			return err
		}
	}
	return nil
}

func Run() int {
	config, _ := LoadConfig()
	sources := make(Source, 0, 1+len(config.Folders))

	recentDir, err := GetRecentDir()
	if err != nil {
		return exitError
	}
	shortcuts, err := NewShortcutInfoList(recentDir)
	if err != nil {
		return exitError
	}
	sources = append(sources, GetShortcutTexts(shortcuts))

	for _, folder := range config.Folders {
		shortcuts, err := NewShortcutInfoList(folder)
		if err != nil {
			continue
		}
		sources = append(sources, GetShortcutTexts(shortcuts))
	}

	selected, err := RunFF(sources)
	if err != nil {
		return exitError
	}
	err = RunApp(selected)
	if err != nil {
		return exitError
	}
	return exitOK
}
