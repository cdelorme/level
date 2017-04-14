package main

import (
	"fmt"
	"sync"
	"time"
)

var printf = fmt.Printf

type Stats struct {
	sync.RWMutex
	start  time.Time
	keys   []string
	values []int64
}

func (s *Stats) init() {
	s.start = time.Now()
}

func (s *Stats) Add(k string, v int) {
	s.Lock()
	defer s.Unlock()
	for i, _ := range s.keys {
		if s.keys[i] == k {
			s.values[i] += int64(v)
			return
		}
	}
	s.keys = append(s.keys, k)
	s.values = append(s.values, int64(v))
}

func (s *Stats) Summary() {
	s.RLock()
	defer s.RUnlock()
	for i, _ := range s.keys {
		printf("%s: %d\n", s.keys[i], s.values[i])
	}
	printf("Total Execution Time: %s\n", time.Since(s.start))
}
