package level6

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/crc32"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

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
	Summary
	MaxParallelism int
	Logger         logger
	Path           string
	Delete         bool
	Move           string
	Summarize      bool
	MaxSize        int64
	Excludes       []string
	Files          map[int64][]File
	Duplicates     map[string][]File
}

func (level6 *Level6) Walk(path string, file os.FileInfo, err error) error {
	if file != nil && file.Mode().IsRegular() {
		f := File{Size: file.Size(), Path: path}
		level6.Summary.Files = level6.Summary.Files + 1

		if level6.MaxSize > 0 && f.Size >= level6.MaxSize {
			return err
		}

		for i, _ := range level6.Excludes {
			if strings.Contains(strings.ToLower(path), level6.Excludes[i]) {
				return err
			}
		}

		if _, ok := level6.Files[f.Size]; !ok {
			level6.Files[f.Size] = make([]File, 0)
		}
		level6.Files[f.Size] = append(level6.Files[f.Size], f)
	}
	return err
}

func (level6 *Level6) HashAndCompare() {
	crc32Sizes := make(chan int64)
	crc32Hashes := make(chan int64)
	var crc32Hashing sync.WaitGroup
	var crc32HashCount sync.WaitGroup

	crc32Hashing.Add(level6.MaxParallelism)
	for i := 0; i < level6.MaxParallelism; i++ {
		go func(num int) {
			defer crc32Hashing.Done()
			hash := crc32.New(nil)

			for size := range crc32Sizes {
				level6.Logger.Debug("channel %d, for file size %d", num, size)
				for i, _ := range level6.Files[size] {
					content, err := ioutil.ReadFile(level6.Files[size][i].Path)
					if err != nil {
						level6.Logger.Error("failed to parse file %s, %s", level6.Files[size][i].Path, err)
						continue
					}

					hash.Write(content)
					level6.Files[size][i].Hash = hex.EncodeToString(hash.Sum(nil))
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
			level6.Summary.Crc32Hashes = level6.Summary.Crc32Hashes + 1
		}
	}()

	for size, _ := range level6.Files {
		if len(level6.Files[size]) > 1 {
			crc32Sizes <- size
		}
	}

	close(crc32Sizes)
	crc32Hashing.Wait()
	close(crc32Hashes)
	crc32HashCount.Wait()

	sizes := make(chan int64)
	var comparison sync.WaitGroup
	duplicates := make(chan map[string][]File)

	sha256Hashes := make(chan int64)
	sha256Duplicates := make(chan int64)
	var sha256Counting sync.WaitGroup
	var sha256DuplicateCounting sync.WaitGroup

	comparison.Add(level6.MaxParallelism)
	for i := 0; i < level6.MaxParallelism; i++ {
		go func() {
			defer comparison.Done()
			hash := sha256.New()

			for size := range sizes {
				crc32Dups := make(map[string][]File)
				sha256Dups := make(map[string][]File)

				for i, _ := range level6.Files[size] {
					if _, ok := crc32Dups[level6.Files[size][i].Hash]; !ok {
						crc32Dups[level6.Files[size][i].Hash] = make([]File, 0)

						for d := i + 1; d < len(level6.Files[size]); d++ {
							if level6.Files[size][i].Hash == level6.Files[size][d].Hash {
								crc32Dups[level6.Files[size][i].Hash] = append(crc32Dups[level6.Files[size][i].Hash], level6.Files[size][d])
							}
						}

						if len(crc32Dups[level6.Files[size][i].Hash]) > 0 {
							crc32Dups[level6.Files[size][i].Hash] = append(crc32Dups[level6.Files[size][i].Hash], level6.Files[size][i])
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
							sha256Dups[h] = make([]File, 0)

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
			level6.Summary.Sha256Hashes = level6.Summary.Sha256Hashes + 1
		}
	}()

	sha256DuplicateCounting.Add(1)
	go func() {
		defer sha256DuplicateCounting.Done()
		for dups := range duplicates {
			for hash, files := range dups {
				if len(dups[hash]) > 0 {
					if _, ok := level6.Duplicates[hash]; !ok {
						level6.Duplicates[hash] = make([]File, 0)
					}
					level6.Duplicates[hash] = append(level6.Duplicates[hash], files...)
					level6.Summary.Duplicates = level6.Summary.Duplicates + int64(len(dups[hash]))
				}
			}
		}
	}()

	for size, _ := range level6.Files {
		if len(level6.Files[size]) > 1 {
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
}

func (level6 *Level6) Finish() {
	if level6.Move != "" {
		if ok, _ := exists(level6.Move); !ok {
			err := os.MkdirAll(level6.Move, 0740)
			if err != nil {
				level6.Logger.Error("Failed to make dir files, %s", err)
			}
		}

		hashes := make(chan string)
		var moving sync.WaitGroup
		moves := make(chan int64)
		var moveCount sync.WaitGroup

		moving.Add(level6.MaxParallelism)
		for i := 0; i < level6.MaxParallelism; i++ {
			go func() {
				defer moving.Done()
				for hash := range hashes {
					for i := 0; i < len(level6.Duplicates[hash])-1; i++ {
						mv := filepath.Join(level6.Move, strings.TrimPrefix(level6.Duplicates[hash][i].Path, level6.Path))
						if err := os.MkdirAll(filepath.Dir(mv), 0740); err != nil {
							level6.Logger.Error("failed to create containing folder %s", filepath.Dir(mv))
						}
						if err := os.Rename(level6.Duplicates[hash][i].Path, mv); err != nil {
							level6.Logger.Error("failed to move %s to %s, %s", level6.Duplicates[hash][i].Path, mv, err)
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
				level6.Summary.Moves = level6.Summary.Moves + 1
			}
		}()

		for hash, _ := range level6.Duplicates {
			hashes <- hash
		}

		close(hashes)
		moving.Wait()
		close(moves)
		moveCount.Wait()

	} else if level6.Delete {
		hashes := make(chan string)
		var deleting sync.WaitGroup
		deletes := make(chan int64)
		var deleteCount sync.WaitGroup

		deleting.Add(level6.MaxParallelism)
		for i := 0; i < level6.MaxParallelism; i++ {
			go func() {
				defer deleting.Done()
				for hash := range hashes {
					for i := 0; i < len(level6.Duplicates[hash])-1; i++ {
						err := os.Remove(level6.Duplicates[hash][i].Path)
						if err != nil {
							level6.Logger.Error("failed to delete file: %s, %s", level6.Duplicates[hash][i].Path, err)
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
				level6.Summary.Deletes = level6.Summary.Deletes + 1
			}
		}()

		for hash, _ := range level6.Duplicates {
			hashes <- hash
		}

		close(hashes)
		deleting.Wait()
		close(deletes)
		deleteCount.Wait()
	}
}
