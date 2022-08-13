package lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
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

type Tmp [][]Shortcut

func RunFF(source []string, query string) (string, error) {
	ff, err := exec.LookPath("peco")
	if err != nil {
		return "", err
	}

	var cmd *exec.Cmd
	if query == "" {
		cmd = exec.Command(ff)
	} else {
		cmd = exec.Command(ff, "--query", query)
	}
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
	fileName := ""
	if strings.HasPrefix(path, folderPrefix) {
		fileName = strings.TrimSpace(strings.Replace(path, folderPrefix, "", -1))
	} else if strings.HasPrefix(path, filePrefix) {
		fileName = strings.TrimSpace(strings.Replace(path, filePrefix, "", -1))
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

func Run() error {
	var cache []Shortcut
	if _, err := os.Stat(CacheDir()); !os.IsNotExist(err) {
		cache, err = NewShortcutList(CacheDir(), Cache)
		if err != nil {
			return err
		}
	}
	config, _ := LoadConfig()
	recentDir, err := GetRecentDir()
	if err != nil {
		return err
	}
	recent, err := NewShortcutList(recentDir, Recent)
	if err != nil {
		return err
	}

	users := make(Tmp, 0, 1+len(config.Folders))
	var userLen int
	for _, folder := range config.Folders {
		user, err := NewShortcutList(folder, User)
		if err != nil {
			continue
		}
		users = append(users, user)
		userLen += len(user)
	}

	shortCutInfos := make([]Shortcut, 0, len(cache)+len(recent)+userLen)
	shortCutInfos = append(shortCutInfos, cache...)
	shortCutInfos = append(shortCutInfos, recent...)
	for _, u := range users {
		shortCutInfos = append(shortCutInfos, u...)
	}
	sort.Slice(shortCutInfos, func(i, j int) bool {
		return shortCutInfos[i].ModTime.After(shortCutInfos[j].ModTime)
	})

	unique := make(map[string]Shortcut)
	texts := make([]string, 0, len(shortCutInfos))
	for _, s := range shortCutInfos {
		t := s.Text()
		if _, ok := unique[t]; ok {
			continue
		}
		unique[t] = s
		texts = append(texts, t)
	}

	query := ""
	for {
		selected, err := RunFF(texts, query)
		if err != nil {
			query = "すいませんもう一度お願いします (Ctrl+uでクリア)"
			continue
		}
		if selected == "" {
			continue
		}
		query = ""
		err = RunApp(selected)
		if err != nil {
			return fmt.Errorf("RunAPP: selected [%s], %w", selected, err)
		}

		if s, ok := unique[selected]; ok {
			switch s.Org {
			case Recent:
				if err := CacheLink(s.Path); err != nil {
					return err
				}
			case Cache:
				if err := TouchLink(s.Path); err != nil {
					return err
				}
			}
		}

	}
	return nil
}
