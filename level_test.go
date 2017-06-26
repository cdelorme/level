package level

// Coverage omissions include any file system activity that is
// difficult to test with complex abstractions.

import (
	"io/ioutil"
	"os"
	"testing"
)

type mockLogger struct{}

func (l *mockLogger) Error(string, ...interface{}) {}
func (l *mockLogger) Info(string, ...interface{})  {}
func (l *mockLogger) Debug(string, ...interface{}) {}

func TestSixStats(*testing.T) {
	s := Six{L: &mockLogger{}}
	s.Stats(ioutil.Discard)
}

func TestSixFiltered(t *testing.T) {
	s := Six{L: &mockLogger{}}
	if n := len(s.Filtered()); n != 0 {
		t.Fatalf("Failed to get 0 results from filtered, instead got %d", n)
	}
}

func TestSixDelete(*testing.T) {
	s := Six{L: &mockLogger{}}
	s.Delete()
}

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
	fileThree, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileThree.Write([]byte("test duplacate")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileThree.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	fileFour, err := ioutil.TempFile(dir, "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileFour.Write([]byte("now for something completely different")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileFour.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	fileFive, err := ioutil.TempFile(dir, "level.ignore")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if _, err := fileFive.Write([]byte("this is nothing")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := fileFive.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	s := Six{L: &mockLogger{}, Input: dir, Excludes: ".ignore"}
	s.LastOrder()
	s.Delete()
}
