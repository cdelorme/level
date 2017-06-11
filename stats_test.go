package level

import (
	"testing"
)

func TestStats(t *testing.T) {
	s := &Stats{}
	s.Add("Key", 1)
	s.Add("Key", 1)
	t.Logf("Stats as String:\n%s", s.String())
	t.Logf("Stats as Json:\n%s", s.Json())
	s.Reset()
	if len(s.keys) != 0 {
		t.Errorf("reset failed to clear old values...")
		t.FailNow()
	}
}
