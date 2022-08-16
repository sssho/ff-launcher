package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
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

func doMemprof(path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	runtime.GC()    // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var debug = flag.String("debug", "", "do not run ff")

func main() {
	flag.Parse()
	cpuprof := *cpuprofile != ""
	memprof := *memprofile != ""
	if cpuprof {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		// defer pprof.StopCPUProfile()
	}
	err := lib.Run(cpuprof || memprof || *debug != "")
	if err != nil {
		_ = writeLog(err)
		os.Exit(1)
	}
	if memprof {
		doMemprof(*memprofile)
	}
	if cpuprof {
		pprof.StopCPUProfile()
	}
	os.Exit(0)
}
