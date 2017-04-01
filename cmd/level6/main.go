package main

import (
	"os"
	"runtime/pprof"

	"github.com/cdelorme/glog"
	"github.com/cdelorme/gonf"
	"github.com/cdelorme/level6"
)

var exit = os.Exit
var create = os.Create
var startp = pprof.StartCPUProfile
var stopp = pprof.StopCPUProfile

type executor interface {
	Execute() error
}

type stats interface {
	Summary()
}

func configure() (executor, stats) {
	cwd, _ := os.Getwd()
	s := &Stats{}
	s.init()

	l6 := &level6.Level6{
		Input:  cwd,
		Stats:  s,
		Logger: &glog.Logger{},
	}

	g := &gonf.Config{}
	g.Description("file deduplication program")
	g.Target(l6)
	g.Add("input", "input path to scan", "LEVEL6_INPUT", "-i:", "--input")
	g.Add("move", "move duplicates to a the given path", "LEVEL6_MOVE", "-m:", "--move")
	g.Add("test", "test run do nothing but print actions", "LEVEL6_TEST", "-t", "--test")
	g.Add("excludes", "comma-delimited patterns to exclude", "LEVEL6_EXCLUDES", "-e:", "--excludes")
	g.Example("-i ~/")
	g.Example("-m ~/dups -i ~/")
	g.Load()

	return l6, s
}

func main() {
	if profile := os.Getenv("GO_PROFILE"); len(profile) > 0 {
		f, _ := create(profile)
		startp(f)
		defer stopp()
	}

	l6, s := configure()
	e := l6.Execute()
	s.Summary()
	if e != nil {
		exit(1)
	}
}
