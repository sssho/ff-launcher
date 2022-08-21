package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/sssho/ffl/lib"
)

func writeLog(logMessage error) error {
	tmpdir := os.Getenv("TEMP")
	logdir := path.Join(tmpdir, "ffl_log")

	now := time.Now()
	logfile := path.Join(logdir, fmt.Sprintf("%s.txt", now.Format("20060102_150405")))

	err := os.MkdirAll(logdir, 0777)
	if err != nil {
		return err
	}
	file, err := os.Create(logfile)
	if err != nil {
		return err
	}
	defer file.Close()

	logger := log.New(file, "log: ", log.Lshortfile)
	logger.Print(logMessage)

	return nil
}

func run() int {
	if err := lib.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		time.Sleep(time.Second * 10)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
