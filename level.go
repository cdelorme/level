// This package provides abstract structures and functions to help coordinate
// a variety of file deduplication methods.
//
// Each of the structures is independently usable, but Six is the primary tool.
package level

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
)

// A minimum logger interface with two levels.
type Logger interface {
	Error(string, ...interface{})
	Debug(string, ...interface{})
}

// An efficient buffered computation for crc32 without memory problems.
func GetBufferedCrc32(path string) (string, error) {
	hasher := crc32.New(nil)
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()
	if _, err := io.Copy(hasher, in); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// An efficient buffered computation for sha256 without memory problems.
func GetBufferedSha256(path string) (string, error) {
	hasher := sha256.New()
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()
	if _, err := io.Copy(hasher, in); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// A buffered byte-comparison function that accepts the paths to two files,
// makes sure they are not hardlinked, then reads them byte by byte in small
// chunks to compare without high memory consumption.
func BufferedByteComparison(one, two string) (bool, error) {
	if one == two {
		return false, nil
	}
	inOne, err := os.Open(one)
	if err != nil {
		return false, err
	}
	defer inOne.Close()
	inTwo, err := os.Open(two)
	if err != nil {
		return false, err
	}
	defer inTwo.Close()

	statOne, err := inOne.Stat()
	if err != nil {
		return false, err
	}
	statTwo, err := inTwo.Stat()
	if err != nil {
		return false, err
	}
	if os.SameFile(statOne, statTwo) {
		return false, nil
	}

	bufferOne, bufferTwo := make([]byte, 32*1024), make([]byte, 32*1024)
	readerOne, readerTwo := bufio.NewReader(inOne), bufio.NewReader(inTwo)
	for {
		_, errOne := readerOne.Read(bufferOne)
		_, errTwo := readerTwo.Read(bufferTwo)
		if errOne == io.EOF || io.EOF == errTwo {
			if errOne == errTwo {
				break
			}
			return false, nil
		}
		if bytes.Compare(bufferOne, bufferTwo) != 0 {
			return false, nil
		}
	}
	return true, nil
}

// A buffered line-by-line comparison that checks whether every line within one
// exists within two as well.  Any failure to read results in an error, and any
// mismatched line will result in false.
func BufferedLineCheck(one, two string) (bool, error) {

	// @todo: perform buffered comparison of each line
	// on first mismatch return false

	return false, nil
}
