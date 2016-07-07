package level6

import (
	"fmt"
	"sync"
	"time"
)

var sprintf = fmt.Sprintf

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

func (self *stats) Summary() string {
	self.RLock()
	defer self.RUnlock()
	var response string
	for k, v := range self.fields {
		response += sprintf("%s: %d\n", k, v)
	}
	response += sprintf("Total Execution Time: %s", time.Since(self.start))
	return response
}
