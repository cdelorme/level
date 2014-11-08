package main

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"
)

func main() {

	// prepare level6 /w logger and empty maps
	level6 := Level6{
		Logger:     log.Logger{Level: log.Error},
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
	appOptions.Flag("quiet", "silence all output", "-q", "--quiet")
	appOptions.Flag("json", "output in json", "-j", "--json")
	appOptions.Flag("verbose", "verbose event output", "-v", "--verbose")
	appOptions.Flag("profile", "create cpu profile", "-p", "--profile")
	appOptions.Example("level6 -p ~/")
	appOptions.Example("level6 -d -p ~/")
	appOptions.Example("level6 -m ~/dups -p ~/")
	flags := appOptions.Parse()

	// apply flags
	level6.Path, _ = maps.String(&flags, cwd, "path")
	level6.Move, _ = maps.String(&flags, level6.Move, "move")
	level6.Json, _ = maps.Bool(&flags, level6.Json, "json")
	level6.Delete, _ = maps.Bool(&flags, level6.Delete, "delete")
	level6.Logger.Silent, _ = maps.Bool(&flags, level6.Logger.Silent, "quiet")
	if ok, _ := maps.Bool(&flags, false, "verbose"); ok {
		level6.Logger.Level = log.Debug
	}

	// if quiet is set but not delete or move then exit
	if level6.Logger.Silent && !level6.Delete && level6.Move == "" {
		level6.Logger.Error("quiet is set but not delete or move, exiting...")
		return
	}

	// profiling
	if profile, _ := maps.String(&flags, "", "profile"); profile != "" {
		f, _ := os.Create(profile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
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

	// @todo implement image, video, and audio comparison algorithms

	// print out results
	level6.Print()
}
