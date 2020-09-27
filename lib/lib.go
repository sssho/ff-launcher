package lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	exitOK    = 0
	exitError = 1
)

const (
	folderPrefix = "[folder]"
	filePrefix   = "[file  ]"
)

type Origin int

const (
	Recent Origin = iota
	Cache
	User
)

func GetRecentDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`%s\AppData\Roaming\Microsoft\Windows\Recent`, home), nil
}

func CacheDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return ".ffl"
	}
	return filepath.Join(filepath.Dir(exePath), ".ffl")
}

type Tmp [][]ShortcutInfo

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

func CacheLink(path string) error {
	cacheDir := CacheDir()
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err := os.Mkdir(cacheDir, os.ModeDir)
		if err != nil {
			return err
		}
	}

	ofile, err := os.Create(filepath.Join(cacheDir, filepath.Base(path)))
	if err != nil {
		return err
	}
	defer ofile.Close()
	ifile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer ifile.Close()
	_, err = io.Copy(ofile, ifile)
	if err != nil {
		return err
	}
	return nil
}

func TouchLink(path string) error {
	if err := os.Chtimes(path, time.Now(), time.Now()); err != nil {
		return err
	}
	return nil
}

func Run() int {
	var cache []ShortcutInfo
	if _, err := os.Stat(CacheDir()); !os.IsNotExist(err) {
		cache, err = NewShortcutInfoList(CacheDir(), Cache)
		if err != nil {
			return exitError
		}
	}
	config, _ := LoadConfig()
	recentDir, err := GetRecentDir()
	if err != nil {
		return exitError
	}
	recent, err := NewShortcutInfoList(recentDir, Recent)
	if err != nil {
		return exitError
	}

	users := make(Tmp, 0, 1+len(config.Folders))
	var userLen int
	for _, folder := range config.Folders {
		user, err := NewShortcutInfoList(folder, User)
		if err != nil {
			continue
		}
		users = append(users, user)
		userLen += len(user)
	}

	shortCutInfos := make([]ShortcutInfo, 0, len(cache)+len(recent)+userLen)
	shortCutInfos = append(shortCutInfos, cache...)
	shortCutInfos = append(shortCutInfos, recent...)
	for _, u := range users {
		shortCutInfos = append(shortCutInfos, u...)
	}

	unique := make(map[string]ShortcutInfo)
	texts := make([]string, 0, len(shortCutInfos))
	for _, s := range shortCutInfos {
		t := s.Text()
		if _, ok := unique[t]; ok {
			continue
		}
		unique[t] = s
		texts = append(texts, t)
	}

	selected, err := RunFF(texts)
	if err != nil {
		return exitError
	}
	err = RunApp(selected)
	if err != nil {
		return exitError
	}

	if s, ok := unique[selected]; ok {
		switch s.Org {
		case Recent:
			if err := CacheLink(s.Path); err != nil {
				return exitError
			}
		case Cache:
			if err := TouchLink(s.Path); err != nil {
				return exitError
			}
		}
	}
	return exitOK
}
