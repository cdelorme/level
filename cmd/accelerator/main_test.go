package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestPlacebo(*testing.T) {}

func TestMain(*testing.T) {
	print = func(...interface{}) (int, error) { return 0, nil }
	stdout = ioutil.Discard

	// test output version
	os.Args = []string{"-ti", "%"}
	main()

	// actually run delete
	os.Args = []string{"-i", "%"}
	main()
}
