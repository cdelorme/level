// This package provides a utility that scans files and checks for duplicates.
package level

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	// @note: testing horrible solution...
	// "time"
)

const statsTotalFilesScanned = "Total Files Scanned"
const statsTotalFilesCollected = "Total Files Collected"
const statsTotalFileComparisons = "Total File Comparisons"
const statsTotalDuplicateFiles = "Total Duplicate Files"
const statsFilesFlaggedForDeletion = "Files Flagged for Deletion"
const statsFilesDeleted = "Files Deleted"
const statsFoldersDeletedDuringCleanup = "Folders Deleted During Cleanup"
const maxFileDescriptors = 512 // based on C standard library

// A minimum logger interface with three severities.
type Logger interface {
	Error(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
}

// A minimum stats interface for adding to collected metrics.
type Stats interface {
	Add(string, int) int
}

// An abstraction for deduplication logic, with a minimal interface.
type Six struct {
	Input    string `json:"input,omitempty"`
	Excludes string `json:"excludes,omitempty"`
	Test     bool   `json:"test,omitempty"`
	L        Logger `json:"-"`
	S        Stats  `json:"-"`

	running     bool
	excludes    []string
	filesBySize map[int64][]string `json:"-"`
	duplicates  [][]string
	filtered    []string
}

func (Six) comparison(one, two string) (bool, error) {
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
	bufferOne, bufferTwo := make([]byte, 4*1024), make([]byte, 4*1024)
	readerOne, readerTwo := bufio.NewReader(inOne), bufio.NewReader(inTwo)
	for {
		bytesOne, errOne := readerOne.Read(bufferOne)
		bytesTwo, errTwo := readerTwo.Read(bufferTwo)
		if bytesOne != bytesTwo || errOne != nil || errTwo != nil {
			if errOne == io.EOF && errOne == errTwo {
				break
			} else if errOne != nil && errOne == io.EOF {
				return false, errOne
			} else if errTwo != nil && errTwo != io.EOF {
				return false, errTwo
			}
			return false, nil
		} else if bytes.Compare(bufferOne, bufferTwo) != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (s *Six) recursion(files []*fileContainer) {
	// one option, slower to release
	// for i := range files {
	// 	defer files[i].f.Close()
	// }
	s.S.Add(statsTotalFileComparisons, len(files)-1)
	var deeper []*fileContainer
	var fc *fileContainer
	for {
		bytesOne, errOne := files[0].r.Read(files[0].b)
		if errOne != nil && errOne != io.EOF {
			s.L.Error("%s", errOne)
			break
		}
		for i := 1; i < len(files); i++ {
			bytesTwo, errTwo := files[i].r.Read(files[i].b)
			if errTwo != errOne || bytesTwo != bytesOne || bytes.Compare(files[0].b, files[i].b) != 0 {
				fc, files = files[i], append(files[:i], files[i+1:]...)
				i--
				if errTwo != nil && errTwo != io.EOF {
					s.L.Error("%s", errTwo)
					continue
				}
				fc.f.Seek(0, 0)
				deeper = append(deeper, fc)
			}
		}
		if errOne == io.EOF {
			break
		}
	}
	if len(files) > 1 {
		var group []string
		for i := range files {
			group = append(group, files[i].f.Name())
		}
		s.duplicates = append(s.duplicates, group)
		s.S.Add(statsTotalDuplicateFiles, len(group))
	}

	// // potentially faster and more direct, gives system time while next subgroup is compared
	// for i := range files {
	// 	files[i].f.Close()
	// }

	if len(deeper) > 1 {
		s.recursion(deeper)
	}

	// // potentially faster and more direct, gives system time while next subgroup is compared
	// // however it may also be redundant as depth continues to add to the process
	// for i := range deeper {
	// 	deeper[i].f.Close()
	// }
}

func (s *Six) data() {
	for size := range s.filesBySize {
		if len(s.filesBySize[size]) < 2 {
			continue
		}

		set := make([]string, len(s.filesBySize[size]))
		copy(set, s.filesBySize[size])

		if len(set) <= maxFileDescriptors {
			var files []*fileContainer
			for i := range set {
				fc, e := newFileContainer(set[i])
				if e != nil {
					s.L.Error("%s", e)
					continue
				}
				files = append(files, fc)
			}
			if len(files) < 2 {
				continue
			}
			s.recursion(files)

			// @note: terrible solution but one which helped avoid the dreaded
			// too many open files on debian linux, at the expense of a sleep
			// after every group is compared, without any real semblance of
			// safety due to the expectation of a synchronous Close on files...
			for i := range files {
				files[i].f.Close()
			}
			// time.Sleep(time.Millisecond * 100)

			continue
		}

		for j := 0; j < len(set)-1; j++ {
			group := []string{}
			for k := j + 1; k < len(set); k++ {
				s.S.Add(statsTotalFileComparisons, 1)
				match, e := s.comparison(set[j], set[k])
				if e != nil {
					s.L.Error("%s", e)
					continue
				}
				if match {
					group = append(group, set[k])
					set = append(set[:k], set[k+1:]...)
					k--
				}
			}
			if len(group) > 0 {
				s.S.Add(statsTotalDuplicateFiles, len(group))
				s.duplicates = append(s.duplicates, append([]string{set[j]}, group...))
			}
		}
	}
}

func (s *Six) filter() {
	frequency := map[string]int{}
	for j := range s.duplicates {
		for k := range s.duplicates[j] {
			d := filepath.Dir(s.duplicates[j][k])
			if _, ok := frequency[d]; !ok {
				frequency[d] = 1
			} else {
				frequency[d] += 1
			}
		}
	}
	for i := range s.duplicates {
		sort.Slice(s.duplicates[i], func(j, k int) bool {
			return strings.Count(s.duplicates[i][j], string(filepath.Separator)) < strings.Count(s.duplicates[i][k], string(filepath.Separator)) || (strings.Count(s.duplicates[i][j], string(filepath.Separator)) == strings.Count(s.duplicates[i][k], string(filepath.Separator)) && frequency[filepath.Dir(s.duplicates[i][j])] < frequency[filepath.Dir(s.duplicates[i][k])])
		})
		s.filtered = append(s.filtered, s.duplicates[i][1:]...)
	}
}

func (s *Six) walk(filePath string, f os.FileInfo, e error) error {
	s.S.Add(statsTotalFilesScanned, 1)
	if e != nil {
		s.L.Error(e.Error())
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
	s.S.Add(statsTotalFilesCollected, 1)
	if _, ok := s.filesBySize[f.Size()]; !ok {
		s.filesBySize[f.Size()] = make([]string, 0)
	}
	s.filesBySize[f.Size()] = append(s.filesBySize[f.Size()], filePath)
	return nil
}

// Returns the duplicates marked for deletion.
func (s *Six) Filtered() []string {
	return s.filtered
}

// Iterate filtered files to delete each, and attempt to clear any empty
// parent folders recursively.
func (s *Six) Delete() {
	s.S.Add(statsFilesFlaggedForDeletion, len(s.filtered))
	for i := range s.filtered {
		d := s.filtered[i]
		if err := os.Remove(d); err != nil {
			s.L.Error("%s", err)
			continue
		}
		s.S.Add(statsFilesDeleted, 1)
		for os.Remove(filepath.Dir(d)) == nil {
			s.S.Add(statsFoldersDeletedDuringCleanup, 1)
			d = filepath.Dir(d)
		}
	}
}

// Initializes the metrics system, which sets the start time and clears data.
//
// Ensures the input path is both absolute and clean, parses the supplied
// excludes, and initializes private maps and slices clearing any former data.
//
// Uses path/filepath.WalkFunc to iterate all files in the input path, and
// discards any zero-size files, symbolic links, or files matching the list
// of case-sensitive excludes.  It groups the remaining files by size.
//
// Any errors encountered while walking the file system will be logged and
// then discarded so the program may continue.
//
// Iterates each set of files grouped by size, and two at a time will be
// checked using os.SameFile to discard hard-links, and then buffered
// byte-by-byte comparison.
//
// The buffered comparison offers early termination, making it a faster
// solution than hash checks.  Additionally, the code is written to work
// with the possibility of multiple duplicate groups of the same size.
//
// Files with matching data are put into an unnamed group and appended
// to the slice of duplicates.
//
// Finally it sorts the groups of duplicates, using a weighted score by depth
// and then by recurrence of parent path.  The file with the lowest score in
// the group will be kept, and the rest are appended to a single dimensional
// slice, which can be requested by Filtered and is used by Delete.
func (s *Six) LastOrder() {
	s.Input, _ = filepath.Abs(s.Input)
	for _, v := range strings.Split(s.Excludes, ",") {
		if v != "" {
			s.excludes = append(s.excludes, v)
		}
	}
	s.filesBySize = map[int64][]string{}
	s.duplicates = [][]string{}
	s.filtered = []string{}
	s.L.Info("%#v", s)
	filepath.Walk(s.Input, s.walk)
	s.data()
	s.filter()
}
