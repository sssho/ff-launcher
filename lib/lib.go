package lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	exitOK    = 0
	exitError = 1
)

const (
	folderPrefix = "[folder]"
	filePrefix   = "[file  ]"
)

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
