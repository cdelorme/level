package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"
)

func main() {

	// prepare level6 /w logger and empty maps
	level6 := Level6{
		Logger:     log.Logger{Level: log.INFO},
		Files:      make(map[int64][]File),
		Duplicates: make(map[string][]File),
	}

	// optimize concurrent processing
	level6.MaxParallelism = runtime.NumCPU()
	runtime.GOMAXPROCS(level6.MaxParallelism)

	// get current directory
	cwd, _ := os.Getwd()

	// prepare cli options
	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("path", "path to begin scanning", "-p", "--path")
	appOptions.Flag("delete", "delete duplicate files", "-d", "--delete")
	appOptions.Flag("move", "move files to supplied path", "-m", "--move")
	appOptions.Flag("verbose", "verbose event output", "-v", "--verbose")
	appOptions.Flag("quiet", "silence all output", "-q", "--quiet")
	o := appOptions.Parse()

	// apply flags
	level6.Path, _ = maps.String(&o, cwd, "path")
	level6.Logger.Silent, _ = maps.Bool(&o, false, "quiet")
	level6.Delete, _ = maps.Bool(&o, false, "delete")
	level6.Move, _ = maps.String(&o, "", "move")
	if ok, _ := maps.Bool(&o, false, "verbose"); ok {
		level6.Logger.Level = log.DEBUG
	}

	// if quiet is set but not delete or move, exit
	if level6.Logger.Silent && !level6.Delete && level6.Move == "" {
		level6.Logger.Error("quiet is set but not delete or move, exiting...")
		return
	}

	// print initial level6 state
	level6.Logger.Debug("initial application state: %+v", level6)

	// build list of files grouped by size
	if err := filepath.Walk(level6.Path, level6.Walk); err != nil {
		level6.Logger.Error("failed to walk directory: %s", err)
	}
	level6.Logger.Debug("files: %+v", level6.Files)

	level6.GenerateHashes()
	level6.Logger.Debug("files /w hashes: %+v", level6.Files)

	// async compare
	level6.CompareHashes()
	level6.Logger.Debug("duplicates: %+v", level6.Duplicates)

	// print out contents as pretty json (readable, and programmatically parsable)
	if !level6.Logger.Silent {
		out, err := json.MarshalIndent(level6.Duplicates, "", "    ")
		if err == nil {
			fmt.Println(string(out))
		}
	}

	// handle move or delete logic
	if level6.Move != "" {
		err := os.Mkdir(level6.Move, 0740)
		if err != nil {
			level6.Logger.Error("Failed to make dir files, %s", err)
		}
		for hash, _ := range level6.Duplicates {
			for i := 0; i < len(level6.Duplicates[hash])-1; i++ {
				d := filepath.Join(level6.Move, hash)
				if err := os.Mkdir(d, 0740); err != nil {
					level6.Logger.Error("failed to create containing folder %s", d)
				}
				mv := filepath.Join(d, strconv.Itoa(i+1)+"-"+filepath.Base(level6.Duplicates[hash][i].Path))
				if err := os.Rename(level6.Duplicates[hash][i].Path, mv); err != nil {
					level6.Logger.Error("failed to move %s to %s, %s", level6.Duplicates[hash][i].Path, mv, err)
				}
			}
		}
	} else if level6.Delete {
		for hash, _ := range level6.Duplicates {
			for i := 0; i < len(level6.Duplicates[hash])-1; i++ {
				err := os.Remove(level6.Duplicates[hash][i].Path)
				if err != nil {
					level6.Logger.Error("failed to delete file: %s, %s", level6.Duplicates[hash][i].Path, err)
				}
			}
		}
	}

	// @todo
	// - mkdir for hash folders for duplicates during mv
	//
	// - test move & delete on large sets of data
	// - test build & execution on Windows
}
