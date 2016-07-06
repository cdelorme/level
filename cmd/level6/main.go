package main

import (
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"

	"github.com/cdelorme/level6"
)

func main() {

	l6 := level6.Level6{
		Logger:     &log.Logger{},
		Files:      make(map[int64][]level6.File),
		Duplicates: make(map[string][]level6.File),
		Summary:    level6.Summary{Start: time.Now()},
		Excludes:   []string{"/."},
	}

	cwd, _ := os.Getwd()

	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("path", "path to begin scanning", "-p", "--path")
	appOptions.Flag("delete", "delete duplicate files", "-d", "--delete")
	appOptions.Flag("move", "move files to supplied path", "-m", "--move")
	appOptions.Flag("excludes", "comma-delimited patterns to exclude", "-e", "--excludes")
	appOptions.Flag("summarize", "print summary at end of operations", "-s", "--summary")
	appOptions.Flag("profile", "output cpu profile to supplied file path", "--profile")
	appOptions.Example("-p ~/")
	appOptions.Example("-d -p ~/")
	appOptions.Example("-m ~/dups -p ~/")
	flags := appOptions.Parse()

	l6.Path, _ = maps.String(flags, cwd, "path")
	l6.Path, _ = filepath.Abs(l6.Path)
	l6.Delete, _ = maps.Bool(flags, l6.Delete, "delete")
	l6.Move, _ = maps.String(flags, l6.Move, "move")
	l6.Summarize, _ = maps.Bool(flags, l6.Summarize, "summarize")

	if e, err := maps.String(flags, "", "excludes"); err == nil {
		l6.Excludes = append(l6.Excludes, strings.Split(strings.ToLower(e), ",")...)
	}

	if profile, _ := maps.String(flags, "", "profile"); profile != "" {
		f, _ := os.Create(profile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	l6.Logger.Debug("initial application state: %+v", l6)

	if err := filepath.Walk(l6.Path, l6.Walk); err != nil {
		l6.Logger.Error("failed to walk directory: %s", err)
	}

	l6.HashAndCompare()
	l6.Finish()
}
