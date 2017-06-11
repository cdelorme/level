package level

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type mockLogger struct{}

func (l *mockLogger) Error(string, ...interface{}) {}
func (l *mockLogger) Debug(string, ...interface{}) {}

func TestSixDelete(t *testing.T) {
	dir, err := ioutil.TempDir("", "level")
	if err != nil {
		t.Errorf("%s", err)
	}
	fileOne, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileOne.Write([]byte("test deletion")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileOne.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	fileTwo, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileTwo.Write([]byte("test deletion")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileTwo.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	fileThree, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	fileThree.Close()
	os.Remove(fileThree.Name())

	s := Six{L: &mockLogger{}}
	s.Delete([]string{fileOne.Name(), fileTwo.Name(), fileThree.Name()})
}

func TestSixFilter(t *testing.T) {
	data := [][]string{
		[]string{"/path/to/duplicate/one", "/deep/path/to/duplicate/one"},
		[]string{"/path/to/duplicate/two", "/deep/path/to/duplicate/two"},
	}
	s := Six{L: &mockLogger{}}

	o := s.Filter(data)
	if len(o) != 2 || o[0] != "/deep/path/to/duplicate/one" || o[1] != "/deep/path/to/duplicate/two" {
		t.Errorf("unexpected outputs: %v\n", o)
		t.FailNow()
	}
}

// Further tests of Data() require known crc32 and sha256 collissions.
func TestSixData(*testing.T) {}

// Due to the combination of checks we only needed to replicate one scenarion
// however it would not be easy to replicate irregular or non-persmissive files
// in a usable way from the tests.
//
// For the same reason we are unable to force a non-nil error parameter.
func TestSixWalk(t *testing.T) {
	dir, err := ioutil.TempDir("", "level")
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.RemoveAll(dir)
	fileOne, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileOne.Write([]byte("test walk")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileOne.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	fileTwo, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileTwo.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	fileThree, err := ioutil.TempFile(dir, ".ignore")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileThree.Write([]byte("test excludes")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileThree.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	s := &Six{L: &mockLogger{}, FilesBySize: map[int64][]string{}, excludes: []string{".ignore"}}
	filepath.Walk(dir, s.Walk)
	if len(s.FilesBySize) != 1 {
		t.Errorf("failed to identify valid file from walk...")
		t.FailNow()
	}
}

// A test demonstrating an invalid path, test, and delete behavior.
func TestSixLastOrder(t *testing.T) {
	dir, err := ioutil.TempDir("", "level")
	if err != nil {
		t.Errorf("%s", err)
	}
	defer os.RemoveAll(dir)
	fileOne, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileOne.Write([]byte("test duplicate")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileOne.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	fileTwo, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileTwo.Write([]byte("test duplicate")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileTwo.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	s := &Six{L: &mockLogger{}, Input: dir, Excludes: ".ignore", Test: true}
	if err := s.LastOrder(); err != nil {
		t.Logf("failed to run LastOrder: %s\n", err)
		t.FailNow()
	} else if len(s.FilesBySize) != 1 {
		t.Errorf("failed to identify valid file from walk...")
		t.FailNow()
	}

	s.Test = false
	if err := s.LastOrder(); err != nil {
		t.Logf("failed to run LastOrder: %s\n", err)
		t.FailNow()
	} else if len(s.FilesBySize) != 1 {
		t.Errorf("failed to identify valid file from walk...")
		t.FailNow()
	}
	t.Logf("%s\n", s.Json())
}
