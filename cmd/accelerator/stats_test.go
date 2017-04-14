package main

import (
	"testing"
)

func init() {
	printf = func(_ string, _ ...interface{}) (int, error) { return 0, nil }
}

func TestStats(t *testing.T) {
	t.Parallel()
	s := Stats{}
	s.init()
	s.Add("Key", 1)
	s.Add("Key", 1)
	s.Summary()
}
