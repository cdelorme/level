package level

import (
	"fmt"
	"sync"
	"time"
)

var start = time.Now()

// A mutex protected subsystem to safely collect statistics during runtime and
// print a humanly readable summary.
type Stats struct {
	mu     sync.RWMutex
	keys   []string
	values []int
}

// Add locks and checks for the key to update the value by what is supplied,
// otherwise it adds a new key and sets the value to what is supplied.
func (s *Stats) Add(k string, v int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, _ := range s.keys {
		if s.keys[i] == k {
			s.values[i] += v
			return
		}
	}
	s.keys = append(s.keys, k)
	s.values = append(s.values, v)
}

// Collects all fields in-order with the final execution time and
// prints them into an indented json structure.
func (s *Stats) Json() string {
	f := "{\n"
	s.mu.Lock()
	for i := range s.keys {
		f += fmt.Sprintf("\t\"%s\": %d,\n", s.keys[i], s.values[i])
	}
	s.mu.Unlock()
	f += fmt.Sprintf("\t\"%s\": \"%s\"\n}", "Total Execution Time", time.Since(start))
	return f
}

// Collect the final values in string format, including execution time,
// and return them.
func (s *Stats) String() string {
	var f string
	s.mu.RLock()
	for i := range s.keys {
		f += fmt.Sprintf("%s: %d\n", s.keys[i], s.values[i])
	}
	s.mu.RUnlock()
	f += fmt.Sprintf("Total Execution Time: %s\n", time.Since(start))
	return f
}
