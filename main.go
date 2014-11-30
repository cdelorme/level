package main

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"

	"github.com/cdelorme/level6/level6"
)

func main() {

	// prepare level6 /w logger and empty maps
	l6 := level6.Level6{
		Logger:     log.Logger{Level: log.Error},
		Files:      make(map[int64][]File),
		Duplicates: make(map[string][]File),
		Summary:    Summary{Start: time.Now()},
		Excludes:   []string{"/."},
	}

	// optimize concurrent processing
	l6.MaxParallelism = runtime.NumCPU()
	runtime.GOMAXPROCS(l6.MaxParallelism)

	// get current directory
	cwd, _ := os.Getwd()

	// prepare cli options
	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("path", "path to begin scanning", "-p", "--path")
	appOptions.Flag("delete", "delete duplicate files", "-d", "--delete")
	appOptions.Flag("move", "move files to supplied path", "-m", "--move")
	appOptions.Flag("max", "maximum file size to hash (in kilobytes)", "--max-size")
	appOptions.Flag("excludes", "comma-delimited patterns to exclude", "-e", "--excludes")
	appOptions.Flag("quiet", "silence all output", "-q", "--quiet")
	appOptions.Flag("json", "output in json", "-j", "--json")
	appOptions.Flag("summarize", "print summary at end of operations", "-s", "--summary")
	appOptions.Flag("verbose", "verbose event output", "-v", "--verbose")
	appOptions.Flag("profile", "output cpu profile to supplied file path", "--profile")
	appOptions.Example("level6 -p ~/")
	appOptions.Example("level6 -d -p ~/")
	appOptions.Example("level6 -m ~/dups -p ~/")
	flags := appOptions.Parse()

	// apply flags
	l6.Path, _ = maps.String(&flags, cwd, "path")
	l6.Path, _ = filepath.Abs(l6.Path)
	l6.Delete, _ = maps.Bool(&flags, l6.Delete, "delete")
	l6.Move, _ = maps.String(&flags, l6.Move, "move")
	if max, _ := maps.Float(&flags, 0, "max"); max > 0 {
		l6.MaxSize = int64(max * 1024)
	}
	l6.Logger.Silent, _ = maps.Bool(&flags, l6.Logger.Silent, "quiet")
	l6.Json, _ = maps.Bool(&flags, l6.Json, "json")
	l6.Summarize, _ = maps.Bool(&flags, l6.Summarize, "summarize")
	if ok, _ := maps.Bool(&flags, false, "verbose"); ok {
		l6.Logger.Level = log.Debug
	}

	// parse excludes
	if e, err := maps.String(&flags, "", "excludes"); err == nil {
		l6.Excludes = append(l6.Excludes, strings.Split(strings.ToLower(e), ",")...)
	}

	// profiling
	if profile, _ := maps.String(&flags, "", "profile"); profile != "" {
		f, _ := os.Create(profile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// if quiet is set but not delete or move then exit
	if l6.Logger.Silent && !l6.Delete && l6.Move == "" {
		l6.Logger.Error("quiet is set but not delete or move, exiting...")
		return
	}

	// print initial level6 state
	l6.Logger.Debug("initial application state: %+v", l6)

	// build list of files grouped by size
	if err := filepath.Walk(l6.Path, l6.Walk); err != nil {
		l6.Logger.Error("failed to walk directory: %s", err)
	}

	// hash and compare async
	l6.HashAndCompare()

	// @todo implement image, video, and audio comparison algorithms

	// finish up by printing results and handling any move or delete operations
	l6.Finish()
}
