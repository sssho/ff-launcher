package lib

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindFromDir(dir string) (hist History, err error) {
	dentries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	hist = make(History, 0, len(dentries))
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
		hist = append(hist, HistItem{filepath.ToSlash(tpath), isdir, finfo.ModTime()})
	}
	return hist, nil
}

func FindFromRecent() (hist History, err error) {
	hist, err = FindFromDir(RecentDir())
	if err != nil {
		return nil, err
	}
	return hist, nil
}

func FindFromUser(folders []string) (hist History, err error) {
	for _, folder := range folders {
		tmpItems, err := FindFromDir(folder)
		if err != nil {
			return nil, err
		}
		hist = append(hist, tmpItems...)
	}
	return hist, nil
}

func FindHistory(config Config) (hist History, err error) {
	var tmpHist History
	if config.EnableHist {
		_ = hist.Load(filepath.Join(config.HistDir, HISTFILE))
	}
	hist.SortByPath()
	if config.EnableRecent {
		tmpHist, err = FindFromRecent()
		if err != nil {
			return History{}, fmt.Errorf("read recent error")
		}
		for i := range tmpHist {
			hist.Merge(tmpHist[i])
		}
	}
	if config.EnableUser {
		tmpHist, err = FindFromUser(config.Folders)
		if err != nil {
			return History{}, fmt.Errorf("read user error")
		}
		for i := range tmpHist {
			hist.Merge(tmpHist[i])
		}
	}
	return hist, nil
}
