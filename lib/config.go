package lib

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Folders []string
}

func LoadConfig() (Config, error) {
	var config Config
	exePath, err := os.Executable()
	if err != nil {
		return config, err
	}
	file, err := os.Open(filepath.Join(filepath.Dir(exePath), "fflconf.json"))
	if err != nil {
		return config, err
	}
	dec := json.NewDecoder(file)
	err = dec.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
