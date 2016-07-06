package main

import (
	"os"
	"testing"

	"github.com/cdelorme/level6"
)

func init() {
	exit = func(_ int) {}
}

func TestPlacebo(_ *testing.T) {}

func TestConfigure(t *testing.T) {
	var ex executor

	// test no parameters
	os.Args = []string{}
	ex = configure()
	if l, e := ex.(*level6.Level6); !e || len(l.Input) == 0 {
		t.FailNow()
	}

	// test with valid parameters
	os.Args = []string{"-t", "-m", "/dups", "-i", "/", "-e", "this,that"}
	ex = configure()
	if l, e := ex.(*level6.Level6); !e || l.Input != "/" || l.Move != "/dups" || !l.Test || l.Excludes != "this,that" {
		t.FailNow()
	}
}

func TestMain(_ *testing.T) {
	os.Args = []string{"-t", "-i", "%"}
	os.Setenv("GO_PROFILE", "/tmp/level6.profile")
	main()
}
