package main

import (
	"os"
	"runtime/pprof"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"

	"github.com/cdelorme/level6"
)

var exit = os.Exit

type executor interface {
	Execute() error
}

func configure() executor {
	cwd, _ := os.Getwd()

	l6 := &level6.Level6{
		Input:  cwd,
		Logger: &log.Logger{},
	}

	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("input", "input path to scan", "-i", "--input")
	appOptions.Flag("move", "move duplicates to a the given path", "-m", "--move")
	appOptions.Flag("test", "test run do nothing but print actions", "-t", "--test")
	appOptions.Flag("excludes", "comma-delimited patterns to exclude", "-e", "--excludes")
	appOptions.Example("-i ~/")
	appOptions.Example("-d -i ~/")
	appOptions.Example("-m ~/dups -i ~/")
	flags := appOptions.Parse()

	maps.To(l6, flags)

	return l6
}

func main() {
	if profile := os.Getenv("GO_PROFILE"); len(profile) > 0 {
		f, _ := os.Create(profile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	l6 := configure()

	if e := l6.Execute(); e != nil {
		exit(1)
	}
}
