package level6

import "os"

type Audio struct {
	stats stats
	l     logger
	err   error
}

func (self *Audio) Stats(s stats) {
	self.stats = s
}

func (self *Audio) Logger(l logger) {
	self.l = l
}

func (self *Audio) Walk(path string, f os.FileInfo, err error) error {
	// @todo: filter files by supported extensions/mimetypes
	return err
}

func (self *Audio) Execute() ([][]string, error) {
	duplicates := [][]string{}
	return duplicates, self.err
}
