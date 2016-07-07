package level6

import (
	"sync"
	"time"
)

type stats struct {
	sync.RWMutex
	start  time.Time
	fields map[string]int64
}

func (self *stats) init() {
	self.Lock()
	defer self.Unlock()
	self.start = time.Now()
	self.fields = make(map[string]int64)
}

func (self *stats) append(key string, value int64) {
	self.Lock()
	defer self.Unlock()
	self.fields[key] += value
}

func (self *stats) summary() {
	self.RLock()
	defer self.RUnlock()
	for k, v := range self.fields {
		printf("%s: %d\n", k, v)
	}
	printf("Total Execution Time: %s\n", time.Since(self.start).String())
}
