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

	"github.com/cdelorme/level6/l6"
)

func main() {

	// prepare level6 /w logger and empty maps
	level6 := l6.Level6{
		Logger:     log.Logger{Severity: log.Error},
		Files:      make(map[int64][]l6.File),
		Duplicates: make(map[string][]l6.File),
		Summary:    l6.Summary{Start: time.Now()},
		Excludes:   []string{"/."},
	}
	level6.Logger.Color()

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
	appOptions.Flag("max", "maximum file size to hash (in kilobytes)", "--max-size")
	appOptions.Flag("excludes", "comma-delimited patterns to exclude", "-e", "--excludes")
	appOptions.Flag("quiet", "silence all output", "-q", "--quiet")
	appOptions.Flag("json", "output in json", "-j", "--json")
	appOptions.Flag("summarize", "print summary at end of operations", "-s", "--summary")
	appOptions.Flag("verbose", "verbose event output", "-v", "--verbose")
	appOptions.Flag("profile", "output cpu profile to supplied file path", "--profile")
	appOptions.Example("-p ~/")
	appOptions.Example("-d -p ~/")
	appOptions.Example("-m ~/dups -p ~/")
	flags := appOptions.Parse()

	// apply flags
	level6.Path, _ = maps.String(flags, cwd, "path")
	level6.Path, _ = filepath.Abs(level6.Path)
	level6.Delete, _ = maps.Bool(flags, level6.Delete, "delete")
	level6.Move, _ = maps.String(flags, level6.Move, "move")
	if max, _ := maps.Float(flags, 0, "max"); max > 0 {
		level6.MaxSize = int64(max * 1024)
	}
	level6.Logger.Silent, _ = maps.Bool(flags, level6.Logger.Silent, "quiet")
	level6.Json, _ = maps.Bool(flags, level6.Json, "json")
	level6.Summarize, _ = maps.Bool(flags, level6.Summarize, "summarize")
	if ok, _ := maps.Bool(flags, false, "verbose"); ok {
		level6.Logger.Severity = log.Debug
	}

	// parse excludes
	if e, err := maps.String(flags, "", "excludes"); err == nil {
		level6.Excludes = append(level6.Excludes, strings.Split(strings.ToLower(e), ",")...)
	}

	// profiling
	if profile, _ := maps.String(flags, "", "profile"); profile != "" {
		f, _ := os.Create(profile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// if quiet is set, but no action such as summary delete or move then exit
	if level6.Logger.Silent && !level6.Summarize && !level6.Delete && level6.Move == "" {
		level6.Logger.Error("quiet is set but not delete or move, exiting...")
		return
	}

	// print initial level6 state
	level6.Logger.Debug("initial application state: %+v", level6)

	// build list of files grouped by size
	if err := filepath.Walk(level6.Path, level6.Walk); err != nil {
		level6.Logger.Error("failed to walk directory: %s", err)
	}

	// hash and compare async
	level6.HashAndCompare()

	// @todo implement image, video, and audio comparison algorithms

	// finish up by printing results and handling any move or delete operations
	level6.Finish()
}
