package level6

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
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

type Level6 struct {
	stats
	sync.Mutex
	Logger logger

	Input    string `json:"input,omitempty"`
	Move     string `json:"move,omitempty"`
	Test     bool   `json:"test,omitempty"`
	Excludes string `json:"excludes,omitempty"`

	err        error
	excludes   []string
	files      map[int64][]file
	duplicates map[string][]file
}

func (self *Level6) copy(in, out string) error {
	if e := os.MkdirAll(filepath.Dir(out), 0740); e != nil {
		return e
	} else if e := os.Rename(in, out); e != nil {
		i, ierr := os.Open(in)
		if ierr != nil {
			return ierr
		}
		defer i.Close()
		o, oerr := os.Open(out)
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

func (self *Level6) exists(path string) (bool, error) {
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
		self.stats.append("Files Processed", 1)
	}
	return err
}

func (self *Level6) compare() error {
	crc32Sizes := make(chan int64)
	var crc32Hashing sync.WaitGroup

	crc32Hashing.Add(8)
	for i := 0; i < 8; i++ {
		go func(num int) {
			defer crc32Hashing.Done()
			hash := crc32.New(nil)

			for size := range crc32Sizes {
				for i, _ := range self.files[size] {
					in, e := os.Open(self.files[size][i].Path)
					if e != nil {
						self.Logger.Error("failed to open file %s, %s", self.files[size][i].Path, e)
						self.error(e)
						continue
					}
					defer in.Close()

					if _, err := io.Copy(hash, in); err != nil {
						self.Logger.Error("failed to hash file %s, %s", self.files[size][i].Path, err)
						self.error(err)
						continue
					}
					self.files[size][i].Hash = hex.EncodeToString(hash.Sum(nil))
					self.stats.append("CRC32 Hashes Created", 1)
					hash.Reset()
				}
			}
		}(i)
	}

	for size, _ := range self.files {
		if len(self.files[size]) > 1 {
			crc32Sizes <- size
		}
	}

	close(crc32Sizes)
	crc32Hashing.Wait()
	sizes := make(chan int64)
	var comparison sync.WaitGroup
	duplicates := make(chan map[string][]file)
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
						in, e := os.Open(crc32Dups[i][f].Path)
						if e != nil {
							self.Logger.Error("failed to open file %s, %s", crc32Dups[i][f].Path, e)
							self.error(e)
							continue
						}
						defer in.Close()

						if _, err := io.Copy(hash, in); err != nil {
							self.Logger.Error("failed to hash file %s, %s", crc32Dups[i][f].Path, err)
							self.error(err)
							continue
						}
						crc32Dups[i][f].Hash = hex.EncodeToString(hash.Sum(nil))
						self.stats.append("SHA256 Hashes Created", 1)
						hash.Reset()
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
					self.stats.append("Duplicates Found", int64(len(dups[hash])))
				}
			}
		}
	}()

	for size, _ := range self.files {
		if len(self.files[size]) > 1 {
			sizes <- size
		}
	}
	close(sizes)
	comparison.Wait()
	close(duplicates)
	sha256DuplicateCounting.Wait()
	return self.err
}

func (self *Level6) error(err error) {
	self.Lock()
	defer self.Unlock()
	self.err = err
}

func (self *Level6) move() error {
	if ok, _ := self.exists(self.Move); !ok {
		if e := os.MkdirAll(self.Move, 0740); e != nil {
			self.Logger.Error("Failed to make dir files, %s", e)
			return e
		}
	}

	hashes := make(chan string)
	var moving sync.WaitGroup

	moving.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			defer moving.Done()
			for hash := range hashes {
				for i := 0; i < len(self.duplicates[hash])-1; i++ {
					mv := filepath.Join(self.Move, strings.TrimPrefix(self.duplicates[hash][i].Path, self.Input))
					if self.Test {
						printf("moving %s\n", mv)
					} else if e := self.copy(self.duplicates[hash][i].Path, mv); e != nil {
						self.Logger.Error("failed to move %s to %s, %s", self.duplicates[hash][i].Path, mv, e)
						self.error(e)
					} else {
						self.stats.append("Files Moved", 1)
					}
				}
			}
		}()
	}

	for hash, _ := range self.duplicates {
		hashes <- hash
	}

	close(hashes)
	moving.Wait()
	return self.err
}

func (self *Level6) delete() error {
	hashes := make(chan string)
	var deleting sync.WaitGroup

	deleting.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			defer deleting.Done()
			for hash := range hashes {
				for i := 0; i < len(self.duplicates[hash])-1; i++ {
					if self.Test {
						printf("deleting %s\n", self.duplicates[hash][i].Path)
					} else {
						if e := os.Remove(self.duplicates[hash][i].Path); e != nil {
							self.Logger.Error("failed to delete file: %s, %s", self.duplicates[hash][i].Path, e)
							self.error(e)
						} else {
							self.stats.append("Deleted Files", 1)
						}
					}
				}
			}
		}()
	}

	for hash, _ := range self.duplicates {
		hashes <- hash
	}

	close(hashes)
	deleting.Wait()
	return self.err
}

func (self *Level6) Execute() error {
	self.stats.init()
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

	if len(self.Move) > 0 {
		return self.move()
	}
	return self.delete()
}
