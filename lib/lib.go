package lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	folderPrefix = "[folder]"
	filePrefix   = "[file  ]"
)

func RecentDir() string {
	return filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Recent")
}

func SelectByFF(r io.Reader, query string, prompt string) (string, error) {
	ff, err := exec.LookPath("peco")
	if err != nil {
		return "", err
	}
	args := []string{}
	if query != "" {
		args = append(args, []string{"--query", query}...)
	}
	if prompt != "" {
		args = append(args, []string{"--prompt", prompt}...)
	}
	cmd := exec.Cmd{
		Path: ff,
		Args: args,
		SysProcAttr: &syscall.SysProcAttr{
			CreationFlags:    0x10, // CREATE_NEW_CONSOLE,
			NoInheritHandles: false,
		},
	}
	cmd.Stdin = r
	cmd.Stderr = os.Stderr
	result, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return filepath.FromSlash(strings.TrimSpace(string(result))), nil
}

func RunApp(path string) error {
	fileName := ""
	if strings.HasPrefix(path, folderPrefix) {
		fileName = strings.TrimSpace(strings.Replace(path, folderPrefix, "", -1))
	} else if strings.HasPrefix(path, filePrefix) {
		fileName = strings.TrimSpace(strings.Replace(path, filePrefix, "", -1))
	} else {
		fileName = path
	}
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    false,
		CmdLine:       fmt.Sprintf(` /C start "pseudo" %s`, fileName),
		CreationFlags: 0,
	}

	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}
