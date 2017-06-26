package level

import (
	"bytes"
	"strings"
	"testing"
)

func TestStats(t *testing.T) {
	s := &Stats{}
	var b bytes.Buffer
	if s.Add("Key", 1) != 1 {
		t.Fatal("expected value to be stored and returned...")
	} else if s.Add("Key", 1) != 2 {
		t.Fatal("expected value to be incremented and returned...")
	} else if s.Print(&b); strings.Count(b.String(), "\n") != 2 {
		t.Fatal("expected exactly one key...")
	} else if s.Duration().Nanoseconds() <= 0 {
		t.Fatal("expected execution duration greater than a single nanosecond...")
	}
	b.Reset()
	s.Reset()
	if s.Print(&b); strings.Count(b.String(), "\n") != 0 {
		t.Fatal("reset failed to clear old values...")
	}
}
