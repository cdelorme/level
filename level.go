package level

import (
	"io"
	"os"
	"path/filepath"
)

func copy(in, out string) error {
	if e := os.MkdirAll(filepath.Dir(out), 0740); e != nil {
		return e
	} else if e := os.Rename(in, out); e != nil {
		i, ierr := os.OpenFile(in, os.O_RDWR, 0777)
		if ierr != nil {
			return ierr
		}
		defer i.Close()
		o, oerr := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0777)
		if oerr != nil {
			return oerr
		}
		defer o.Close()

		if _, cerr := io.Copy(o, i); cerr != nil {
			return cerr
		}
		i.Close()
		if derr := os.Remove(in); derr != nil {
			return derr
		}
	}
	return nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
