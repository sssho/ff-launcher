package lib

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Folders      []string
	CacheDir     string
	EnableRecent bool
	EnableUser   bool
	EnableCache  bool
	DefaultQuery string
	OneShot      bool
	CachePath    string
}

func LoadConfig() (Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return Config{nil, "", true, true, true, "", false, ""}, err
	}
	file, err := os.Open(filepath.Join(filepath.Dir(exePath), "fflconf.json"))
	if err != nil {
		return Config{nil, "", true, true, true, "", false, ""}, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	var config Config
	err = dec.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
