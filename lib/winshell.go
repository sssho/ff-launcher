package lib

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
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

func GetShortcutInfo(path string, w *WscriptShell) (string, string, error) {
	shortcut, err := oleutil.CallMethod(w.Wshell, "CreateShortcut", path)
	if err != nil {
		return "", "", fmt.Errorf("createshortcut error!{%s}: %w", path, err)
	}
	shortcutDispath := shortcut.ToIDispatch()

	targetPath, err := shortcutDispath.GetProperty("TargetPath")
	if err != nil {
		return "", "", fmt.Errorf("targetpath error!: %w", err)
	}

	args, err := shortcutDispath.GetProperty("Arguments")
	if err != nil {
		return "", "", fmt.Errorf("arguments error!: %w", err)
	}
	return targetPath.ToString(), args.ToString(), nil
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
