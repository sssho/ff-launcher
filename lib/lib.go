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

type Shortcut struct {
	Path  string
	TPath string
	Args  string
}

func GetShortcutList(dir string) ([]Shortcut, error) {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	w, err := NewWscriptShell()
	if err != nil {
		return nil, err
	}
	defer w.Release()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	shortcuts := make([]Shortcut, 0, len(files))
	for _, file := range files {
		tpath, args, err := w.GetShortcutInfo(filepath.Join(dir, file.Name()))
		if err != nil {
			continue
		}
		if tpath == "" {
			continue
		}
		shortcuts = append(shortcuts, Shortcut{file.Name(), tpath, args})
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
		if checkDuplicate[s.TPath] {
			continue
		} else {
			checkDuplicate[s.TPath] = true
		}
		isDir, err := isDir(s.TPath)
		if err != nil {
			continue
		}
		if isDir {
			texts = append(texts, fmt.Sprintf("%s %s", folderPrefix, s.TPath))
		} else {
			d := filepath.Dir(s.TPath)
			if !checkDuplicate[d] {
				texts = append(texts, fmt.Sprintf("%s %s", folderPrefix, d))
				checkDuplicate[d] = true
			}
			texts = append(texts, fmt.Sprintf("%s %s %s", filePrefix, s.TPath, s.Args))
		}
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

func RunFF(source []string) (string, error) {
	ff, err := exec.LookPath("peco")
	if err != nil {
		return "", err
	}

	cmd := exec.Command(ff)
	cmd.Stderr = os.Stderr
	in, _ := cmd.StdinPipe()
	go func() {
		defer in.Close()

		for _, s := range source {
			io.WriteString(in, s+"\n")
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
	recentDir, err := GetRecentDir()
	if err != nil {
		return exitError
	}
	shortcuts, err := GetShortcutList(recentDir)
	if err != nil {
		return exitError
	}
	selected, err := RunFF(GetShortcutTexts(shortcuts))
	if err != nil {
		return exitError
	}
	err = RunApp(selected)
	if err != nil {
		return exitError
	}
	return exitOK
}
