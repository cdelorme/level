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

func (self *Stats) init() {
	self.start = time.Now()
}

func (self *Stats) Add(key string, v int) {
	self.Lock()
	defer self.Unlock()
	for i, _ := range self.keys {
		if self.keys[i] == key {
			self.values[i] += int64(v)
			return
		}
	}
	self.keys = append(self.keys, key)
	self.values = append(self.values, int64(v))
}

func (self *Stats) Summary() {
	self.RLock()
	defer self.RUnlock()
	for i, _ := range self.keys {
		printf("%s: %d\n", self.keys[i], self.values[i])
	}
	printf("Total Execution Time: %s\n", time.Since(self.start))
}
