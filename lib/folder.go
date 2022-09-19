package lib

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func GetOpenedFolderPaths() (paths []string, err error) {
	// https://learn.microsoft.com/en-us/windows/win32/shell/shell
	// https://learn.microsoft.com/en-us/windows/win32/shell/shellwindows
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	defer ole.CoUninitialize()

	oleShellObject, err := oleutil.CreateObject("Shell.Application")
	if err != nil {
		return nil, err
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, err
	}
	defer wshell.Release()
	wins, err := oleutil.CallMethod(wshell, "Windows")
	if err != nil {
		return nil, err
	}
	winsDispath := wins.ToIDispatch()
	counti, err := winsDispath.GetProperty("Count")
	if err != nil {
		return nil, fmt.Errorf("counti error!: %w", err)
	}
	count, ok := (counti.Value()).(int32)
	if !ok {
		return nil, errors.New("count error")
	}
	for i := 0; i < int(count); i++ {
		ie, err := oleutil.CallMethod(winsDispath, "Item", i)
		if err != nil {
			return nil, err
		}
		ieDispath := ie.ToIDispatch()
		path, err := ieDispath.GetProperty("LocationURL")
		if err != nil {
			return nil, fmt.Errorf("ie error!: %w", err)
		}
		paths = append(paths, strings.Replace(path.ToString(), "file:///", "", 1))
	}
	return paths, nil
}
