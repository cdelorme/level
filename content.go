package level6

import "os"

type Content struct {
	stats stats
	l     logger
	err   error
}

func (self *Content) Stats(s stats) {
	self.stats = s
}

func (self *Content) Logger(l logger) {
	self.l = l
}

func (self *Content) Walk(path string, f os.FileInfo, err error) error {
	// @todo: filter by text mimetype and keep track of size
	// @note: sort by size with simple expectation that only larger files can
	//        contain smaller files worth of contents, except if newlines exceed
	return err
}

func (self *Content) Execute() ([][]string, error) {
	duplicates := [][]string{}
	return duplicates, self.err
}
