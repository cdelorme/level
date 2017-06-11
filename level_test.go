package level

import (
	"io/ioutil"
	"os"
	"testing"
)

// Verify that we can read and generate a crc32 when provided a file path.
//
// There appears to be no way to force io.Copy to throw an error without
// manipulating the inputs.
func TestGetBufferedCrc32(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte("level crc32 test")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := tmpfile.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	crc, err := GetBufferedCrc32(tmpfile.Name())
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	} else if crc == "" {
		t.Errorf("failed to generate a crc32...")
		t.FailNow()
	}
	t.Logf("successfully generated a crc32: %s\n", crc)

	tmpfile, err = ioutil.TempFile("", "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	tmpfile.Close()
	os.Remove(tmpfile.Name())
	if _, err := GetBufferedCrc32(tmpfile.Name()); err == nil {
		t.Errorf("failed to receive error from bad file...")
		t.FailNow()
	}
}

// Verify that we can read and generate a sha256 when provided a file path.
//
// There appears to be no way to force io.Copy to throw an error without
// manipulating the inputs.
//
// Testing that the end
func TestGetBufferedSha256(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte("level sha256 test")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if err := tmpfile.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	sha, err := GetBufferedSha256(tmpfile.Name())
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	} else if sha == "" {
		t.Errorf("failed to generate a sha256...")
		t.FailNow()
	}
	t.Logf("successfully generated a sha256: %s\n", sha)

	tmpfile, err = ioutil.TempFile("", "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	tmpfile.Close()
	os.Remove(tmpfile.Name())
	if _, err := GetBufferedSha256(tmpfile.Name()); err == nil {
		t.Errorf("failed to receive error from bad file...")
		t.FailNow()
	}
}

// Verify that we can compare files byte-by-byte and get accurate results.
//
// Verify that os.SameFile works as expected to detect and ignore hard links.
//
// It may not be possible to create a reliable test without direct control over
// the buffers that will validate situations where contents are similar but one
// file ends early.
//
// All obvious error checking is not tested, which would occur when permissions
// are not allowed or the file does not exist.
func TestBufferedByteComparison(t *testing.T) {
	fileOne, err := ioutil.TempFile("", "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	defer os.Remove(fileOne.Name())
	if _, err := fileOne.Write([]byte("level byte test")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	} else if err = fileOne.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	fileTwo, err := ioutil.TempFile("", "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	defer os.Remove(fileTwo.Name())
	if _, err := fileTwo.Write([]byte("level byte test")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	} else if err = fileTwo.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if b, e := BufferedByteComparison(fileOne.Name(), fileTwo.Name()); e != nil {
		t.Errorf("%s", e)
		t.FailNow()
	} else if !b {
		t.Errorf("failed to correctly compare bytes...\n")
		t.FailNow()
	}

	fileThree, err := ioutil.TempFile("", "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	defer os.Remove(fileThree.Name())
	if _, err := fileThree.Write([]byte("level bytes test")); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	} else if err = fileThree.Close(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if b, e := BufferedByteComparison(fileOne.Name(), fileThree.Name()); e != nil {
		t.Errorf("%s", e)
		t.FailNow()
	} else if b {
		t.Errorf("failed to detect difference in bytes...\n")
		t.FailNow()
	}

	fileFour, err := ioutil.TempFile("", "level")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	fileFour.Close()
	os.Remove(fileFour.Name())
	if err := os.Link(fileThree.Name(), fileFour.Name()); err != nil {
		t.Errorf("failed to create link to test SameFile: %s\n", err)
		t.FailNow()
	}
	if b, e := BufferedByteComparison(fileOne.Name(), fileThree.Name()); e != nil {
		t.Errorf("%s", e)
		t.FailNow()
	} else if b {
		t.Errorf("failed to detect hard link...\n")
		t.FailNow()
	}
}
