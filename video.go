package level6

import "os"

type Video struct {
	stats stats
	l     logger
	err   error
}

func (self *Video) Stats(s stats) {
	self.stats = s
}

func (self *Video) Logger(l logger) {
	self.l = l
}

func (self *Video) Walk(path string, f os.FileInfo, err error) error {
	// @todo: filter files by supported extensions/mimetypes
	return err
}

func (self *Video) Execute() ([][]string, error) {
	duplicates := [][]string{}
	return duplicates, self.err
}
