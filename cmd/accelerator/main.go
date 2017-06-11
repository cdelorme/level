package main

import (
	"os"

	"github.com/cdelorme/glog"
	"github.com/cdelorme/gonf"
	"github.com/cdelorme/level"
)

var exit = os.Exit
var create = os.Create

func main() {
	cwd, _ := os.Getwd()

	l := &glog.Logger{}
	six := &level.Six{Input: cwd, L: l}

	config := &gonf.Config{}
	config.Target(six)
	config.Description("file deduplication program")
	config.Add("input", "input path to scan (defaults to current directory)", "ACCEL_INPUT", "-i:", "--input")
	config.Add("excludes", "comma-delimited patterns to exclude", "ACCEL_EXCLUDES", "-e:", "--excludes")
	config.Add("test", "test run do nothing but print actions", "ACCEL_TEST", "-t", "--test")
	config.Example("-ti ~/")
	config.Example("-i ~/")
	config.Load()

	l.Info("%#v", six)

	if six.LastOrder() != nil {
		exit(1)
	}
}
