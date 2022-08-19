package lib

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindFromDir(dir string) (items []HistItem, err error) {
	dentries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	items = make([]HistItem, 0, len(dentries))
	for _, dentry := range dentries {
		path := filepath.Join(dir, dentry.Name())
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()
		tpath, isdir, _, err := ResolveShortcut(f)
		if err != nil {
			continue
		}
		finfo, err := dentry.Info()
		if err != nil {
			continue
		}
		_, err = os.Stat(tpath)
		if err != nil {
			continue
		}
		items = append(items, HistItem{tpath, isdir, finfo.ModTime()})
	}
	return items, nil
}

func FindFromRecent() (items []HistItem, err error) {
	recentDir, err := GetRecentDir()
	if err != nil {
		return nil, err
	}
	items, err = FindFromDir(recentDir)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func FindFromUser(folders []string) (items []HistItem, err error) {
	for _, folder := range folders {
		tmpItems, err := FindFromDir(folder)
		if err != nil {
			return nil, err
		}
		items = append(items, tmpItems...)
	}
	return items, nil
}

func FindHistory(config Config) (hist History, err error) {
	var tmpItems []HistItem
	if config.EnableCache {
		_ = hist.Load(config.CachePath)
	}
	if config.EnableRecent {
		tmpItems, err = FindFromRecent()
		if err != nil {
			return History{}, fmt.Errorf("read recent error")
		}
		for _, item := range tmpItems {
			if found := contains(item, hist.items); !found {
				hist.items = append(hist.items, item)
			}
		}
	}
	if config.EnableUser {
		tmpItems, err = FindFromUser(config.Folders)
		if err != nil {
			return History{}, fmt.Errorf("read user error")
		}
		for _, item := range tmpItems {
			if found := contains(item, hist.items); !found {
				hist.items = append(hist.items, item)
			}
		}
	}
	hist.Sort()
	return hist, nil
}
