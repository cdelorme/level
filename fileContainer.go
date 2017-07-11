package level

import (
	"bufio"
	"os"
)

type fileContainer struct {
	f *os.File
	b []byte
	r *bufio.Reader
}

func newFileContainer(file string) (*fileContainer, error) {
	fc := &fileContainer{b: make([]byte, 4*1024)}
	var err error
	if fc.f, err = os.Open(file); err != nil {
		return fc, err
	}
	fc.r = bufio.NewReader(fc.f)
	return fc, nil
}
