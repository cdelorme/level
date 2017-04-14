package level

import (
	"os"
	"path/filepath"
	"strings"
)

type Six struct {
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

func (s *Six) walk(path string, f os.FileInfo, err error) error {
	if f != nil && f.Mode().IsRegular() {
		for i, _ := range s.excludes {
			if strings.Contains(strings.ToLower(path), s.excludes[i]) {
				return err
			}
		}
		return s.steps[s.step].Walk(path, f, err)
	}
	return err
}

func (s *Six) move() {
	if ok, _ := exists(s.Move); !ok {
		if e := os.MkdirAll(s.Move, 0740); e != nil {
			s.Logger.Error("Failed to make dir files, %s", e)
			s.err = e
			return
		}
	}

	for i, _ := range s.duplicates {
		for _, v := range s.duplicates[i] {
			mv := filepath.Join(s.Move, strings.TrimPrefix(v, s.Input))
			if s.Test {
				s.Logger.Info("moving %s\n", mv)
			} else if e := copy(v, mv); e != nil {
				s.Logger.Error("failed to move %s to %s, %s", v, mv, e)
				s.err = e
			} else {
				s.Stats.Add(StatsMoved, 1)
			}
		}
	}
}

func (s *Six) delete() {
	for i, _ := range s.duplicates {
		for _, v := range s.duplicates[i] {
			if s.Test {
				s.Logger.Info("deleting %s\n", v)
			} else if e := os.Remove(v); e != nil {
				s.Logger.Error("failed to delete file: %s, %s", v, e)
				s.err = e
			} else {
				s.Stats.Add(StatsDeleted, 1)
			}
		}
	}
}

func (s *Six) Step(stp step) {
	s.steps = append(s.steps, stp)
}

func (s *Six) Execute() error {
	if s.Logger == nil {
		s.Logger = &nilLogger{}
	}
	if s.Stats == nil {
		s.Stats = &nilStats{}
	}
	s.Input, _ = filepath.Abs(filepath.Clean(s.Input))
	s.excludes = append([]string{"/."}, strings.Split(s.Excludes, ",")...)
	if len(s.Move) > 0 {
		s.Move, _ = filepath.Abs(filepath.Clean(s.Move))
		s.excludes = append(s.excludes, s.Move)
	}
	for i := 0; i < len(s.excludes); i++ {
		if len(s.excludes[i]) == 0 {
			s.excludes = append(s.excludes[:i], s.excludes[i+1:]...)
			i--
		}
	}
	if s.step == 0 {
		s.steps = append(steps, s.steps...)
	} else {
		s.step = 0
	}
	s.Logger.Debug("initial state: %#v", s)

	var stp step
	for s.step, stp = range steps {
		if t, ok := stp.(statCollector); ok {
			t.Stats(s.Stats)
		}
		if l, ok := stp.(loggable); ok {
			l.Logger(s.Logger)
		}
		if err := filepath.Walk(s.Input, s.walk); err != nil {
			s.Logger.Error("failed to walk directory: %s", err)
			s.err = err
			continue
		}
		dups, err := stp.Execute()
		if err != nil {
			s.err = err
		}
		s.duplicates = append(s.duplicates, dups...)
		if len(s.Move) > 0 {
			s.move()
		} else {
			s.delete()
		}
		s.duplicates = [][]string{}
	}

	return s.err
}
