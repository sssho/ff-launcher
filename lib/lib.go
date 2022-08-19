package lib

import (
	"encoding/json"
	"fmt"
	"io"
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

func WriteCache(path string, items []HistItem) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	err = e.Encode(items)
	if err != nil {
		return err
	}
	return nil
}

func ReadCache(path string) (items []HistItem, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	err = d.Decode(&items)
	if err != nil {
		return nil, err
	}
	return items, err
}

func RunFF(source []HistItem, query string) (string, error) {
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

func SortHistItems(h []HistItem) {
	sort.Slice(h, func(i, j int) bool {
		return h[i].lastAccess.After(h[j].lastAccess)
	})
}

func FindFromDir(dir string, org Origin) (items []HistItem, err error) {
	items, err = NewHistItems(dir, org)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func FindFromRecent() (items []HistItem, err error) {
	recentDir, err := GetRecentDir()
	if err != nil {
		return nil, err
	}
	items, err = FindFromDir(recentDir, Recent)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func FindFromUser(folders []string) (items []HistItem, err error) {
	for _, folder := range folders {
		tmpItems, err := FindFromDir(folder, User)
		if err != nil {
			return nil, err
		}
		items = append(items, tmpItems...)
	}
	return items, nil
}

func contains(h HistItem, hitems []HistItem) bool {
	for _, hitem := range hitems {
		if h.path == hitem.path {
			return true
		}
	}
	return false
}

func FindHistItems(config Config) (items []HistItem, err error) {
	var tmpItems []HistItem
	if config.EnableCache {
		items, _ = ReadCache(config.CachePath)
	}
	if config.EnableRecent {
		tmpItems, err = FindFromRecent()
		if err != nil {
			return nil, fmt.Errorf("read recent error")
		}
		for _, item := range tmpItems {
			if found := contains(item, items); !found {
				items = append(items, item)
			}
		}
	}
	if config.EnableUser {
		tmpItems, err = FindFromUser(config.Folders)
		if err != nil {
			return nil, fmt.Errorf("read user error")
		}
		for _, item := range tmpItems {
			if found := contains(item, items); !found {
				items = append(items, item)
			}
		}
	}
	SortHistItems(items)
	return items, nil
}

func Run(debug bool) error {
	config, _ := LoadConfig()
	SetupCache(&config)
	histItems, err := FindHistItems(config)
	if err != nil {
		return err
	}
	// err = WriteCache(config.CachePath, shortcuts.unique)
	// if err != nil {
	// 	return err
	// }
	query := config.DefaultQuery
	if debug {
		for i, s := range histItems {
			fmt.Println(i, s.path)
			// fmt.Printf("%d %+v\n", i, s)
		}
		return err
	}
	for {
		selected, err := RunFF(histItems, query)
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
