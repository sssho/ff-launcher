package lib

import (
	"bytes"
	"fmt"
	"io"
	"log"
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

func SelectByFF(r io.Reader, query string, prompt string) (string, error) {
	ff, err := exec.LookPath("peco")
	if err != nil {
		return "", err
	}
	args := []string{"cmd", "/c", ff}
	if query != "" {
		args = append(args, []string{"--query", query}...)
	}
	if prompt != "" {
		args = append(args, []string{"--prompt", prompt}...)
	}
	cmd := exec.Command("cmd", args...)
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

func Run(debug bool) error {
	config, _ := LoadConfig()
	SetupCache(&config)
	hist, err := FindHistory(config)
	if err != nil {
		return err
	}
	query := config.DefaultQuery
	if debug {
		for i, s := range hist {
			fmt.Println(i, s.path)
			// fmt.Printf("%d %+v\n", i, s)
		}
		// TestHistory()
		err = hist.Save(config.CachePath)
		if err != nil {
			log.Fatal("save error", err)
		}
		return err
	}
	hist.SortByTime()
	var b bytes.Buffer
	for {
		b.Reset()
		for _, h := range hist {
			_, err = fmt.Fprintf(&b, "%s\n", h.path)
			if err != nil {
				return err
			}
		}
		selected, err := SelectByFF(&b, query, "")
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
