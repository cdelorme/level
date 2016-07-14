package level6

import "os"

type Image struct {
	stats stats
	l     logger
	err   error
}

func (self *Image) Stats(s stats) {
	self.stats = s
}

func (self *Image) Logger(l logger) {
	self.l = l
}

func (self *Image) Walk(path string, f os.FileInfo, err error) error {
	// @todo: filter files by supported extensions/mimetypes
	return err
}

func (self *Image) Execute() ([][]string, error) {
	duplicates := [][]string{}
	return duplicates, self.err
}
