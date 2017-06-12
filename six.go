package level

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// A minimum logger interface with two levels.
type Logger interface {
	Error(string, ...interface{})
	Debug(string, ...interface{})
}

// The primary structure that abstracts the logic surrounding deduplication.
//
// Several publicly exposed functions and properties are provided so that each
// can be individually managed.
//
// It requires a valid logger or execution will fail with nil pointer errors.
//
// Since most operations are bound by disk IO nothing is run in parallel.  Any
// attempt to manipulate properties or call functions across goroutines may
// lead to race conditions.
type Six struct {
	Input    string `json:"input,omitempty"`
	Excludes string `json:"excludes,omitempty"`
	Test     bool   `json:"test,omitempty"`

	L           Logger `json:"-"`
	Stats       `json:"-"`
	FilesBySize map[int64][]string `json:"-"`

	err      error
	excludes []string
}

// Delete accepts a single dimensional array of absolute file paths.
//
// An attempt is made to delete each file, and any errors are logged and
// collected to determine the response code of the program.
//
// An attempt is made to cleanup by deleting the parent directory, and
// successful operations will be collected as statistics, but errors will
// be discarded.
func (s *Six) Delete(duplicates []string) {
	s.Stats.Add("Files Flagged for Deletion", len(duplicates))
	for i := range duplicates {
		d := duplicates[i]
		if err := os.Remove(d); err != nil {
			s.L.Error("%s", err)
			s.err = err
			continue
		}
		s.Stats.Add("Files Deleted", 1)
		for os.Remove(filepath.Dir(d)) == nil {
			s.Stats.Add("Folders Deleted During Cleanup", 1)
			d = filepath.Dir(d)
		}
	}
}

// Filter accepts a multi-dimensional array of grouped duplicate files.
//
// It builds a frequency score by path so it can identify common paths with
// duplicates.
//
// Next it sorts favoring depth first, then frequency when depth is equal.
//
// Finally it extracts all but the first element into a single-dimensional
// array of resulting files.
func (s *Six) Filter(duplicates [][]string) []string {
	frequency := map[string]int{}
	for j := range duplicates {
		for k := range duplicates[j] {
			d := filepath.Dir(duplicates[j][k])
			if _, ok := frequency[d]; !ok {
				frequency[d] = 1
			} else {
				frequency[d] += 1
			}
		}
	}
	var final []string
	for i := range duplicates {
		sort.Slice(duplicates[i], func(j, k int) bool {
			return strings.Count(duplicates[i][j], string(filepath.Separator)) < strings.Count(duplicates[i][k], string(filepath.Separator)) || (strings.Count(duplicates[i][j], string(filepath.Separator)) == strings.Count(duplicates[i][k], string(filepath.Separator)) && frequency[filepath.Dir(duplicates[i][j])] < frequency[filepath.Dir(duplicates[i][k])])
		})
		final = append(final, duplicates[i][1:]...)
	}
	return final
}

// Data accepts files grouped by size, which are sequentially compared.
//
// It moves from least to most expensive operations to reduce the cost of
// identifying non-duplicates.
//
// Starting with a buffered crc32 comparison, then a sha256 comparison, and
// finally byte-by-byte comparison which begins with os.SameFile to check and
// ignore hard-links.
//
// The final byte comparison is written under the assumption that a sha256
// conflict is possible, and can produce multiple separate slices.
//
// Any time the number of files remaining in the slice is greater than two it
// is appended to a slice of slices to be returned.
func (s *Six) Data(m map[int64][]string) [][]string {
	delete(m, 0)
	var duplicates [][]string
	for size := range m {
		if len(m[size]) < 2 {
			continue
		}

		filesByCrc := map[string][]string{}
		for i := range m[size] {
			hash, err := GetBufferedCrc32(m[size][i])
			s.Stats.Add("Total Hashes (crc32)", 1)
			if err != nil {
				s.err = err
				s.L.Error("%s", err)
				continue
			}
			if _, ok := filesByCrc[hash]; !ok {
				filesByCrc[hash] = []string{}
			}
			filesByCrc[hash] = append(filesByCrc[hash], m[size][i])
		}

		filesBySha := map[string][]string{}
		for h := range filesByCrc {
			if len(filesByCrc[h]) < 2 {
				continue
			}
			for i := range filesByCrc[h] {
				hash, err := GetBufferedSha256(filesByCrc[h][i])
				s.Stats.Add("Total Hashes (sha256)", 1)
				if err != nil {
					s.err = err
					s.L.Error("%s", err)
					continue
				}

				if _, ok := filesBySha[hash]; !ok {
					filesBySha[hash] = []string{}
				}
				filesBySha[hash] = append(filesBySha[hash], filesByCrc[h][i])
			}
		}

		for _, set := range filesBySha {
			for j := 0; j < len(set)-1; j++ {
				match := []string{}
				for k := j + 1; k < len(set); k++ {
					s.Stats.Add("Total Byte Comparisons", 1)
					m, e := BufferedByteComparison(set[j], set[k])
					if e != nil {
						s.L.Error("%s", e)
						continue
					} else if m {
						match = append(match, set[k])
						set = append(set[:k], set[k+1:]...)
						k--
					}
				}
				if len(match) > 0 {
					s.Stats.Add("Total Duplicate Files", len(match))
					duplicates = append(duplicates, append([]string{set[j]}, match...))
				}
			}
		}
	}

	return duplicates
}

// Implements the path/filepath.Walkfunc format to iterate all files in a
// given directory.
//
// It eliminates zero-size files, symbolic links, and any files that match the
// list of excludes.
//
// Finally it groups files by size, which is a primitive but accurate measure
// provided the block size does not vary when used across multiple hard disks.
//
// To avoid terminating on errors created by inaccessible files, the return
// value is always nil.  However, these errors will be logged and collected
// to determine the response code of the program.
func (s *Six) Walk(filePath string, f os.FileInfo, e error) error {
	s.Stats.Add("Total Files Scanned", 1)
	if e != nil {
		s.L.Error(e.Error())
		s.err = e
	}
	if f == nil || f.IsDir() || !f.Mode().IsRegular() || f.Mode()&os.ModeSymlink != 0 || f.Size() == 0 {
		s.L.Debug("discarding irregular file: %s", filePath)
		return nil
	}
	for i, _ := range s.excludes {
		if strings.Contains(filePath, s.excludes[i]) {
			s.L.Debug("discarding excluded file: %s", filePath)
			return nil
		}
	}
	s.Stats.Add("Total Files Collected", 1)
	if _, ok := s.FilesBySize[f.Size()]; !ok {
		s.FilesBySize[f.Size()] = make([]string, 0)
	}
	s.FilesBySize[f.Size()] = append(s.FilesBySize[f.Size()], filePath)
	return nil
}

// This is core function that coordinates all of the deduplication efforts.
//
// It converts the input path to an absolute path.  Next it builds the excludes
// list, and initializes the files by size map.
//
// It calls Walk() to iterate files in the provided path and build the map of
// files by size.
//
// It calls to Data() with the map, and collects the resulting duplicates.
//
// If running in test mode, it prints every file found, otherwise it runs the
// Delete() function.
//
// Finally it prints a summary to standard out, in json format.
//
// If any errors are encountered during the process, they will be logged and
// collected.  The last error encountered will be returned, otherwise nil.
func (s *Six) LastOrder() error {
	s.Input, _ = filepath.Abs(s.Input)
	for _, v := range strings.Split(s.Excludes, ",") {
		if v != "" {
			s.excludes = append(s.excludes, v)
		}
	}
	s.FilesBySize = map[int64][]string{}
	filepath.Walk(s.Input, s.Walk)
	duplicates := s.Data(s.FilesBySize)
	if s.Test {
		for i := range duplicates {
			fmt.Printf("%v\n", duplicates[i])
		}
	} else {
		s.Delete(s.Filter(duplicates))
	}
	fmt.Println(s.Json())
	return s.err
}
