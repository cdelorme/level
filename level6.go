package level6

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Level6 struct {
	Logger logger `json:"-"`
	Stats  stats  `json:"-"`

	Move     string `json:"move,omitempty"`
	Test     bool   `json:"test,omitempty"`
	Input    string `json:"input,omitempty"`
	Excludes string `json:"excludes,omitempty"`

	err        error
	step       int
	steps      []step
	excludes   []string
	duplicates [][]string
}

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

func (self *Level6) walk(path string, f os.FileInfo, err error) error {
	if f != nil && f.Mode().IsRegular() {
		for i, _ := range self.excludes {
			if strings.Contains(strings.ToLower(path), self.excludes[i]) {
				return err
			}
		}
		return self.steps[self.step].Walk(path, f, err)
	}
	return err
}

func (self *Level6) move() {
	if ok, _ := exists(self.Move); !ok {
		if e := os.MkdirAll(self.Move, 0740); e != nil {
			self.Logger.Error("Failed to make dir files, %s", e)
			self.err = e
			return
		}
	}

	for i, _ := range self.duplicates {
		for _, v := range self.duplicates[i] {
			mv := filepath.Join(self.Move, strings.TrimPrefix(v, self.Input))
			if self.Test {
				self.Logger.Info("moving %s\n", mv)
			} else if e := copy(v, mv); e != nil {
				self.Logger.Error("failed to move %s to %s, %s", v, mv, e)
				self.err = e
			} else {
				self.Stats.Add(StatsMoved, 1)
			}
		}
	}
}

func (self *Level6) delete() {
	for i, _ := range self.duplicates {
		for _, v := range self.duplicates[i] {
			if self.Test {
				self.Logger.Info("deleting %s\n", v)
			} else if e := os.Remove(v); e != nil {
				self.Logger.Error("failed to delete file: %s, %s", v, e)
				self.err = e
			} else {
				self.Stats.Add(StatsDeleted, 1)
			}
		}
	}
}

func (self *Level6) Step(s step) {
	self.steps = append(self.steps, s)
}

func (self *Level6) Execute() error {
	if self.Logger == nil {
		self.Logger = &nilLogger{}
	}
	if self.Stats == nil {
		self.Stats = &nilStats{}
	}
	self.Input, _ = filepath.Abs(filepath.Clean(self.Input))
	self.excludes = append([]string{"/."}, strings.Split(self.Excludes, ",")...)
	if len(self.Move) > 0 {
		self.Move, _ = filepath.Abs(filepath.Clean(self.Move))
		self.excludes = append(self.excludes, self.Move)
	}
	for i := 0; i < len(self.excludes); i++ {
		if len(self.excludes[i]) == 0 {
			self.excludes = append(self.excludes[:i], self.excludes[i+1:]...)
			i--
		}
	}
	if self.step == 0 {
		self.steps = append(steps, self.steps...)
	} else {
		self.step = 0
	}
	self.Logger.Debug("initial state: %#v", self)

	var s step
	for self.step, s = range steps {
		if t, ok := s.(statCollector); ok {
			t.Stats(self.Stats)
		}
		if l, ok := s.(loggable); ok {
			l.Logger(self.Logger)
		}
		if err := filepath.Walk(self.Input, self.walk); err != nil {
			self.Logger.Error("failed to walk directory: %s", err)
			self.err = err
			continue
		}
		dups, err := s.Execute()
		if err != nil {
			self.err = err
		}
		self.duplicates = append(self.duplicates, dups...)
		if len(self.Move) > 0 {
			self.move()
		} else {
			self.delete()
		}
		self.duplicates = [][]string{}
	}

	return self.err
}
