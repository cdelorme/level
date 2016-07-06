package level6

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var printf = fmt.Printf

type logger interface {
	Debug(string, ...interface{})
	Error(string, ...interface{})
	Info(string, ...interface{})
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

type Level6 struct {
	stats
	Logger logger

	Input    string `json:"input,omitempty"`
	Move     string `json:"move,omitempty"`
	Test     bool   `json:"test,omitempty"`
	Excludes string `json:"excludes,omitempty"`

	excludes   []string
	files      map[int64][]file
	duplicates map[string][]file
}

func (self *Level6) walk(path string, f os.FileInfo, err error) error {
	if f != nil && f.Mode().IsRegular() {
		fi := file{Size: f.Size(), Path: path}

		for i, _ := range self.excludes {
			if strings.Contains(strings.ToLower(path), self.excludes[i]) {
				return err
			}
		}

		if _, ok := self.files[fi.Size]; !ok {
			self.files[fi.Size] = make([]file, 0)
		}
		self.files[fi.Size] = append(self.files[fi.Size], fi)
		self.stats.Files++
	}
	return err
}

func (self *Level6) compare() (err error) {
	crc32Sizes := make(chan int64)
	crc32Hashes := make(chan int64)
	var crc32Hashing sync.WaitGroup
	var crc32HashCount sync.WaitGroup

	crc32Hashing.Add(8)
	for i := 0; i < 8; i++ {
		go func(num int) {
			defer crc32Hashing.Done()
			hash := crc32.New(nil)

			for size := range crc32Sizes {
				for i, _ := range self.files[size] {
					content, err := ioutil.ReadFile(self.files[size][i].Path)
					if err != nil {
						self.Logger.Error("failed to parse file %s, %s", self.files[size][i].Path, err)
						continue
					}

					hash.Write(content)
					self.files[size][i].Hash = hex.EncodeToString(hash.Sum(nil))
					hash.Reset()
					crc32Hashes <- 1
				}
			}
		}(i)
	}

	crc32HashCount.Add(1)
	go func() {
		defer crc32HashCount.Done()
		for _ = range crc32Hashes {
			self.stats.Crc32Hashes++
		}
	}()

	for size, _ := range self.files {
		if len(self.files[size]) > 1 {
			crc32Sizes <- size
		}
	}

	close(crc32Sizes)
	crc32Hashing.Wait()
	close(crc32Hashes)
	crc32HashCount.Wait()

	sizes := make(chan int64)
	var comparison sync.WaitGroup
	duplicates := make(chan map[string][]file)

	sha256Hashes := make(chan int64)
	var sha256Counting sync.WaitGroup

	comparison.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			defer comparison.Done()
			hash := sha256.New()

			for size := range sizes {
				crc32Dups := make(map[string][]file)
				sha256Dups := make(map[string][]file)

				for i, _ := range self.files[size] {
					if _, ok := crc32Dups[self.files[size][i].Hash]; !ok {
						crc32Dups[self.files[size][i].Hash] = make([]file, 0)

						for d := i + 1; d < len(self.files[size]); d++ {
							if self.files[size][i].Hash == self.files[size][d].Hash {
								crc32Dups[self.files[size][i].Hash] = append(crc32Dups[self.files[size][i].Hash], self.files[size][d])
							}
						}

						if len(crc32Dups[self.files[size][i].Hash]) > 0 {
							crc32Dups[self.files[size][i].Hash] = append(crc32Dups[self.files[size][i].Hash], self.files[size][i])
						}
					}
				}

				for i, _ := range crc32Dups {
					for f, _ := range crc32Dups[i] {
						content, _ := ioutil.ReadFile(crc32Dups[i][f].Path)

						hash.Write(content)
						crc32Dups[i][f].Hash = hex.EncodeToString(hash.Sum(nil))
						hash.Reset()

						sha256Hashes <- 1
					}
				}

				for h, _ := range crc32Dups {
					for i, _ := range crc32Dups[h] {
						if _, ok := sha256Dups[h]; !ok {
							sha256Dups[h] = make([]file, 0)

							for f := i + 1; f < len(crc32Dups[h]); f++ {

								if crc32Dups[h][i].Hash == crc32Dups[h][f].Hash {
									sha256Dups[h] = append(sha256Dups[h], crc32Dups[h][f])
								}
							}

							if len(sha256Dups[h]) > 0 {
								sha256Dups[h] = append(sha256Dups[h], crc32Dups[h][i])
							}
						}
					}
				}

				if len(sha256Dups) > 0 {
					duplicates <- sha256Dups
				}
			}
		}()
	}

	sha256Counting.Add(1)
	go func() {
		defer sha256Counting.Done()
		for _ = range sha256Hashes {
			self.stats.Sha256Hashes++
		}
	}()

	sha256Duplicates := make(chan int64)
	var sha256DuplicateCounting sync.WaitGroup
	sha256DuplicateCounting.Add(1)
	go func() {
		defer sha256DuplicateCounting.Done()
		for dups := range duplicates {
			for hash, files := range dups {
				if len(dups[hash]) > 0 {
					if _, ok := self.duplicates[hash]; !ok {
						self.duplicates[hash] = make([]file, 0)
					}
					self.duplicates[hash] = append(self.duplicates[hash], files...)
					sha256Duplicates <- int64(len(dups[hash]))
				}
			}
		}
	}()

	sha256DuplicateCounting.Add(1)
	go func() {
		defer sha256DuplicateCounting.Done()
		for v := range sha256Duplicates {
			self.stats.Duplicates += v
		}
	}()

	for size, _ := range self.files {
		if len(self.files[size]) > 1 {
			sizes <- size
		}
	}
	close(sizes)
	comparison.Wait()

	close(sha256Hashes)
	close(sha256Duplicates)
	close(duplicates)
	sha256Counting.Wait()
	sha256DuplicateCounting.Wait()
	return
}

func (self *Level6) finish() (err error) {
	if len(self.Move) > 0 {
		if ok, _ := exists(self.Move); !ok {
			err := os.MkdirAll(self.Move, 0740)
			if err != nil {
				self.Logger.Error("Failed to make dir files, %s", err)
			}
		}

		hashes := make(chan string)
		var moving sync.WaitGroup
		moves := make(chan int64)
		var moveCount sync.WaitGroup

		moving.Add(8)
		for i := 0; i < 8; i++ {
			go func() {
				defer moving.Done()
				for hash := range hashes {
					for i := 0; i < len(self.duplicates[hash])-1; i++ {
						mv := filepath.Join(self.Move, strings.TrimPrefix(self.duplicates[hash][i].Path, self.Input))
						if self.Test {
							printf("moving %s\n", mv)
						} else {
							if err := os.MkdirAll(filepath.Dir(mv), 0740); err != nil {
								self.Logger.Error("failed to create containing folder %s", filepath.Dir(mv))
							}
							if err := os.Rename(self.duplicates[hash][i].Path, mv); err != nil {
								self.Logger.Error("failed to move %s to %s, %s", self.duplicates[hash][i].Path, mv, err)
							}
						}
						moves <- 1
					}
				}
			}()
		}

		moveCount.Add(1)
		go func() {
			defer moveCount.Done()
			for _ = range moves {
			}
		}()

		for hash, _ := range self.duplicates {
			hashes <- hash
		}

		close(hashes)
		moving.Wait()
		close(moves)
		moveCount.Wait()

	} else {
		hashes := make(chan string)
		var deleting sync.WaitGroup
		deletes := make(chan int64)
		var deleteCount sync.WaitGroup

		deleting.Add(8)
		for i := 0; i < 8; i++ {
			go func() {
				defer deleting.Done()
				for hash := range hashes {
					for i := 0; i < len(self.duplicates[hash])-1; i++ {
						if self.Test {
							printf("deleting %s\n", self.duplicates[hash][i].Path)
						} else {
							err := os.Remove(self.duplicates[hash][i].Path)
							if err != nil {
								self.Logger.Error("failed to delete file: %s, %s", self.duplicates[hash][i].Path, err)
							}

						}
						deletes <- 1
					}
				}
			}()
		}

		deleteCount.Add(1)
		go func() {
			defer deleteCount.Done()
			for _ = range deletes {
			}
		}()

		for hash, _ := range self.duplicates {
			hashes <- hash
		}

		close(hashes)
		deleting.Wait()
		close(deletes)
		deleteCount.Wait()
	}
	return
}

func (self *Level6) Execute() error {
	self.Input, _ = filepath.Abs(filepath.Clean(self.Input))
	self.excludes = append([]string{"/."}, strings.Split(self.Excludes, ",")...)
	self.files = make(map[int64][]file)
	self.duplicates = make(map[string][]file)
	if len(self.Move) > 0 {
		self.Move, _ = filepath.Abs(filepath.Clean(self.Move))
		self.excludes = append(self.excludes, self.Move)
	}
	for i := range self.excludes {
		if len(self.excludes[i]) == 0 {
			self.excludes = append(self.excludes[:i], self.excludes[i+1:]...)
		}
	}
	self.Logger.Debug("initial state: %#v", self)

	if err := filepath.Walk(self.Input, self.walk); err != nil {
		self.Logger.Error("failed to walk directory: %s", err)
		return err
	}

	if err := self.compare(); err != nil {
		self.Logger.Error("%s\n", err)
		return err
	}

	self.stats.summary()

	// temporary for testing
	return nil

	if err := self.finish(); err != nil {
		return nil
	}

	return nil
}
