package main

import (
	"os"
	"testing"
)

func TestPlacebo(*testing.T) {}

func TestMain(*testing.T) {
	exit = func(int) {}
	create = func(string) (*os.File, error) { return nil, nil }
	os.Args = []string{"-ti", "%"}
	os.Setenv("GO_PROFILE", "/tmp/accelerator.profile")
	main()
}
