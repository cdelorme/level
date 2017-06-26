package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cdelorme/glog"
	"github.com/cdelorme/gonf"
	"github.com/cdelorme/level"
	"github.com/cdelorme/metrics"
)

var print = fmt.Println
var stdout metrics.Writer = os.Stdout

func main() {
	cwd, _ := os.Getwd()
	stats := metrics.NewStats()
	six := &level.Six{Input: cwd, L: &glog.Logger{}, S: stats}

	config := &gonf.Config{}
	config.Target(six)
	config.Description("file deduplication program")
	config.Add("input", "input path to scan (defaults to current directory)", "ACCEL_INPUT", "-i:", "--input")
	config.Add("excludes", "comma-delimited patterns to exclude", "ACCEL_EXCLUDES", "-e:", "--excludes")
	config.Add("test", "test run do nothing but print actions", "ACCEL_TEST", "-t", "--test")
	config.Example("-ti ~/")
	config.Example("-i ~/")
	config.Load()

	defer stats.Print(stdout)
	six.LastOrder()
	if six.Test {
		d, _ := json.MarshalIndent(six.Filtered(), "", "\t")
		print(string(d))
	} else {
		six.Delete()
	}
}
