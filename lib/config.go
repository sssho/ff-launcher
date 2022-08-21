package lib

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	EnableRecent bool     `json:"enable_recent"`
	EnableUser   bool     `json:"enable_user"`
	EnableHist   bool     `json:"enable_hist"`
	Folders      []string `json:"folders"`
	HistDir      string
}

const HISTFILE = "fflhist.json"

func (c *Config) Load() error {
	path := filepath.Join(os.Getenv("APPDATA"), "ffl", "fflconf.json")
	file, err := os.Open(path)
	if err != nil {
		c.EnableRecent = true
		c.EnableUser = false
		c.EnableHist = false
		c.Folders = nil
		return nil
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	err = dec.Decode(c)
	if err != nil {
		return err
	}
	return nil
}
