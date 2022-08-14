package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
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

func SetupCache(config *Config) {
	name := "ffl_cache.json"
	if _, err := os.Stat(config.CacheDir); !os.IsNotExist(err) {
		config.CachePath = filepath.Join(config.CacheDir, name)
		return
	}
	// default
	defaultDir := filepath.Join(os.Getenv("AppData"), "ffl")
	err := os.Mkdir(defaultDir, 0750)
	if err != nil && !os.IsExist(err) {
		return
	}
	config.CachePath = filepath.Join(defaultDir, name)
}

func WriteCache(path string, s []Shortcut) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	io.WriteString(f, string(b))
	return nil
}

func ReadCache(path string) (s []Shortcut, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var m []Shortcut
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return m, err
}

func RunFF(source *Shortcuts, query string) (string, error) {
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

		for _, s := range source.unique {
			io.WriteString(in, s.Text()+"\n")
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

type Shortcuts struct {
	recent []Shortcut
	cache  []Shortcut
	user   []Shortcut
	merge  []Shortcut
	unique []Shortcut
}

func (s *Shortcuts) Merge() {
	s.merge = make([]Shortcut, 0, len(s.recent)+len(s.cache)+len(s.user))
	s.merge = append(s.merge, s.cache...)
	s.merge = append(s.merge, s.recent...)
	s.merge = append(s.merge, s.user...)
}

func (s *Shortcuts) Sort() {
	sort.Slice(s.merge, func(i, j int) bool {
		return s.merge[i].ModTime.After(s.merge[j].ModTime)
	})
}

func (s *Shortcuts) Unique() {
	s.unique = make([]Shortcut, 0, len(s.merge))
	uni := make(map[string]bool)
	for _, v := range s.merge {
		t := v.Text()
		if _, ok := uni[t]; ok {
			continue
		}
		uni[t] = true
		s.unique = append(s.unique, v)
	}
}

func ReadDir(dir string, org Origin) (shortcuts []Shortcut, err error) {
	sc, err := NewShortcutList(dir, org)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

func FindFromRecent() (shortcuts []Shortcut, err error) {
	recentDir, err := GetRecentDir()
	if err != nil {
		return nil, err
	}
	sc, err := ReadDir(recentDir, Recent)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

func FindFromUser(folders []string) (shortcuts []Shortcut, err error) {
	var sc []Shortcut
	for _, folder := range folders {
		sc_, err := ReadDir(folder, User)
		if err != nil {
			return nil, err
		}
		sc = append(sc, sc_...)
	}
	return sc, nil
}

func FindShortcuts(config Config) (s *Shortcuts, err error) {
	var shortcuts *Shortcuts = &Shortcuts{}
	if config.EnableCache {
		shortcuts.cache, _ = ReadCache(config.CachePath)
		// if err != nil {
		// 	return nil, fmt.Errorf("read cache error")
		// }
	}
	if config.EnableRecent {
		shortcuts.recent, err = FindFromRecent()
		if err != nil {
			return nil, fmt.Errorf("read recent error")
		}
	}
	if config.EnableUser {
		shortcuts.user, err = FindFromUser(config.Folders)
		if err != nil {
			return nil, fmt.Errorf("read user error")
		}
	}
	shortcuts.Merge()
	shortcuts.Sort()
	shortcuts.Unique()

	return shortcuts, nil
}

func Run() error {
	config, _ := LoadConfig()
	SetupCache(&config)
	shortcuts, err := FindShortcuts(config)
	if err != nil {
		return err
	}
	for _, v := range shortcuts.unique {
		fmt.Println(v.Text())
	}
	err = WriteCache(config.CachePath, shortcuts.unique)
	if err != nil {
		return err
	}
	query := config.DefaultQuery
	for {
		selected, err := RunFF(shortcuts, query)
		if err != nil {
			query = "すいませんもう一度お願いします (Ctrl+uでクリア)"
			continue
		}
		if selected == "" {
			continue
		}
		query = config.DefaultQuery
		err = RunApp(selected)
		if err != nil {
			return fmt.Errorf("RunAPP: selected [%s], %w", selected, err)
		}
		if config.OneShot {
			break
		}
	}
	return nil
}
