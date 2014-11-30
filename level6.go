package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cdelorme/go-log"
)

type Level6 struct {
	Summary
	MaxParallelism int
	Logger         log.Logger
	Path           string
	Delete         bool
	Json           bool
	Move           string
	Summarize      bool
	MaxSize        int64
	Files          map[int64][]File
	Duplicates     map[string][]File
}

func (level6 *Level6) Walk(path string, file os.FileInfo, err error) error {
	if file != nil && file.Mode().IsRegular() {
		f := File{Size: file.Size(), Path: path}
		level6.Summary.Files = level6.Summary.Files + 1

		// if the size is above max-size skip file
		if level6.MaxSize > 0 && f.Size >= level6.MaxSize {
			return err
		}

		// if path contains anything in excludes, skip file
		for i, _ := range excludes {
			if strings.Contains(strings.ToLower(path), excludes[i]) {
				return err
			}
		}

		// create missing indexes, then append to that index
		if _, ok := level6.Files[f.Size]; !ok {
			level6.Files[f.Size] = make([]File, 0)
		}
		level6.Files[f.Size] = append(level6.Files[f.Size], f)
	}
	return err
}

func (level6 *Level6) HashAndCompare() {

	// prepare channels and wait groups for async crc32 generation & counting
	crc32Sizes := make(chan int64, level6.MaxParallelism*2)
	crc32Hashes := make(chan int64, level6.MaxParallelism*2)
	var crc32Hashing sync.WaitGroup
	var crc32HashCount sync.WaitGroup

	// create some go routines to hash files in parallel by size
	crc32Hashing.Add(level6.MaxParallelism)
	for i := 0; i < level6.MaxParallelism; i++ {
		go func(num int) {
			defer crc32Hashing.Done()

			// use a single shared hash per-channel
			hash := crc32.New(nil)

			// iterate each supplied size
			for size := range crc32Sizes {
				level6.Logger.Debug("channel %d, for file size %d", num, size)
				for i, _ := range level6.Files[size] {
					content, err := ioutil.ReadFile(level6.Files[size][i].Path)
					if err != nil {
						level6.Logger.Error("failed to parse file %s, %s", level6.Files[size][i].Path, err)
						continue
					}

					// generate hash, add to instance, and reset hash
					hash.Write(content)
					level6.Files[size][i].Hash = hex.EncodeToString(hash.Sum(nil))
					hash.Reset()

					// append to async hash counter
					crc32Hashes <- 1
				}
			}
		}(i)
	}

	// count in parallel
	crc32HashCount.Add(1)
	go func() {
		defer crc32HashCount.Done()
		for _ = range crc32Hashes {
			level6.Summary.Crc32Hashes = level6.Summary.Crc32Hashes + 1
		}
	}()

	// send sets of files to each go routine
	for size, _ := range level6.Files {
		if len(level6.Files[size]) > 1 {
			crc32Sizes <- size
		}
	}

	// close channels and wait for done() calls before moving forward
	close(crc32Sizes)
	crc32Hashing.Wait()
	close(crc32Hashes)
	crc32HashCount.Wait()

	// channels and wait group for async crc32 hash comparison, plus duplicate catching
	sizes := make(chan int64, level6.MaxParallelism*2)
	var comparison sync.WaitGroup
	duplicates := make(chan map[string][]File, level6.MaxParallelism*2)

	// channels and wait groups for counting sha256 hashes and duplicates
	sha256Hashes := make(chan int64, level6.MaxParallelism*2)
	sha256Duplicates := make(chan int64, level6.MaxParallelism*2)
	var sha256Counting sync.WaitGroup
	var sha256DuplicateCounting sync.WaitGroup

	// go routines for async crc32 comparison, sha256 hashing, then sha256 comparison
	comparison.Add(level6.MaxParallelism)
	for i := 0; i < level6.MaxParallelism; i++ {
		go func() {
			defer comparison.Done()

			// shared sha256 hashing component
			hash := sha256.New()

			for size := range sizes {

				// storage for crc32 duplicates and sha256 duplicates
				crc32Dups := make(map[string][]File)
				sha256Dups := make(map[string][]File)

				// iterate crc32 hashes and identify duplicates
				for i, _ := range level6.Files[size] {

					// only create each index once, if it already exists assume we counted it
					if _, ok := crc32Dups[level6.Files[size][i].Hash]; !ok {

						// create a fresh index
						crc32Dups[level6.Files[size][i].Hash] = make([]File, 0)

						// iterate all files after the current index
						for d := i + 1; d < len(level6.Files[size]); d++ {

							// append any duplicates found
							if level6.Files[size][i].Hash == level6.Files[size][d].Hash {
								crc32Dups[level6.Files[size][i].Hash] = append(crc32Dups[level6.Files[size][i].Hash], level6.Files[size][d])
							}
						}

						// append first index if any duplicates were found
						if len(crc32Dups[level6.Files[size][i].Hash]) > 0 {
							crc32Dups[level6.Files[size][i].Hash] = append(crc32Dups[level6.Files[size][i].Hash], level6.Files[size][i])
						}
					}
				}

				// generate sha256 hashes for all items in duplicates
				for i, _ := range crc32Dups {
					for f, _ := range crc32Dups[i] {

						// read file contents (ignore errors since we read it once already)
						content, _ := ioutil.ReadFile(crc32Dups[i][f].Path)

						// apply and reset sha256 hash to duplicate
						hash.Write(content)
						crc32Dups[i][f].Hash = hex.EncodeToString(hash.Sum(nil))
						hash.Reset()

						// add to hash count
						sha256Hashes <- 1
					}
				}

				// compare and move all sha256 duplicates
				for h, _ := range crc32Dups {

					// iterate all files within
					for i, _ := range crc32Dups[h] {

						// only continue if we have not already checked this hash
						if _, ok := sha256Dups[h]; !ok {

							// create index
							sha256Dups[h] = make([]File, 0)

							// compare files at index+1
							for f := i + 1; f < len(crc32Dups[h]); f++ {

								// compare hashes and append duplicates
								if crc32Dups[h][i].Hash == crc32Dups[h][f].Hash {
									sha256Dups[h] = append(sha256Dups[h], crc32Dups[h][f])
								}
							}

							// if duplicates exist by hash, add current record (first-instance)
							if len(sha256Dups[h]) > 0 {
								sha256Dups[h] = append(sha256Dups[h], crc32Dups[h][i])
							}
						}
					}
				}

				// send duplicates for aggregation (append & count)
				if len(sha256Dups) > 0 {
					duplicates <- sha256Dups
				}
			}
		}()
	}

	// count sha256 hashes
	sha256Counting.Add(1)
	go func() {
		defer sha256Counting.Done()
		for _ = range sha256Hashes {
			level6.Summary.Sha256Hashes = level6.Summary.Sha256Hashes + 1
		}
	}()

	// count and capture sha256 duplicates
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

	// begin comparison in parallel
	for size, _ := range level6.Files {
		if len(level6.Files[size]) > 1 {
			sizes <- size
		}
	}

	// finished sending sizes for comparison, now wait for processing to complete
	close(sizes)
	comparison.Wait()

	// close the remaining channels and finish waiting for the counting to complete
	close(sha256Hashes)
	close(sha256Duplicates)
	close(duplicates)
	sha256Counting.Wait()
	sha256DuplicateCounting.Wait()

	// @todo add a high-fidelity option that performs full binary byte-by-byte comparison of two or more files
}

func (level6 *Level6) Finish() {

	// if we are not operating in "quiet" mode, print found items as json or as raw text
	if !level6.Logger.Silent {
		if level6.Json {
			out, err := json.MarshalIndent(level6.Duplicates, "", "    ")
			if err == nil {
				fmt.Println(string(out))
			}
		} else {
			for hash, _ := range level6.Duplicates {
				for i, _ := range level6.Duplicates[hash] {
					fmt.Printf("%s %d %s\n", level6.Duplicates[hash][i].Hash, level6.Duplicates[hash][i].Size, level6.Duplicates[hash][i].Path)
				}
			}
		}
	}

	// determine whether to move or delete found duplicates
	if level6.Move != "" {

		// attempt to create move path if it does not exist
		if ok, _ := exists(level6.Move); !ok {
			err := os.MkdirAll(level6.Move, 0740)
			if err != nil {
				level6.Logger.Error("Failed to make dir files, %s", err)
			}
		}

		// prepare buffered channel and wait group for parallel file renaming
		hashes := make(chan string, level6.MaxParallelism*2)
		var moving sync.WaitGroup

		// prepare bean counters for parallel file renaming
		moves := make(chan int64, level6.MaxParallelism*2)
		var moveCount sync.WaitGroup

		// prepare go routines and add to wait group
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

						// send notification to bean counter
						moves <- 1
					}
				}
			}()
		}

		// count in parallel
		moveCount.Add(1)
		go func() {
			defer moveCount.Done()
			for _ = range moves {
				level6.Summary.Moves = level6.Summary.Moves + 1
			}
		}()

		// send each hash for parallel processing
		for hash, _ := range level6.Duplicates {
			hashes <- hash
		}

		// close channels and wait for done() calls before moving forward
		close(hashes)
		moving.Wait()
		close(moves)
		moveCount.Wait()

	} else if level6.Delete {

		// prepare buffered channel and wait group for parallel file renaming
		hashes := make(chan string, level6.MaxParallelism*2)
		var deleting sync.WaitGroup

		// prepare bean counters for parallel file renaming
		deletes := make(chan int64, level6.MaxParallelism*2)
		var deleteCount sync.WaitGroup

		// prepare go routines and add to wait group
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

						// send notification to bean counter
						deletes <- 1
					}
				}
			}()
		}

		// count in parallel
		deleteCount.Add(1)
		go func() {
			defer deleteCount.Done()
			for _ = range deletes {
				level6.Summary.Deletes = level6.Summary.Deletes + 1
			}
		}()

		// send each hash for parallel processing
		for hash, _ := range level6.Duplicates {
			hashes <- hash
		}

		// close channels and wait for done() calls before moving forward
		close(hashes)
		deleting.Wait()
		close(deletes)
		deleteCount.Wait()
	}

	// determine whether to print summary (ignores "quiet" mode)
	if level6.Summarize {

		// calculation completion time
		level6.Summary.Time = time.Since(level6.Summary.Start)

		// print as json or as raw text
		if level6.Json {
			out, err := json.MarshalIndent(level6.Summary, "", "    ")
			if err == nil {
				fmt.Println(string(out))
			}
		} else {
			fmt.Println("Summary:")
			fmt.Printf("Total files scanned: %d\n", level6.Summary.Files)
			fmt.Printf("Total crc32 hashes generated: %d\n", level6.Summary.Crc32Hashes)
			fmt.Printf("Total sha256 hashes generated: %d\n", level6.Summary.Sha256Hashes)
			fmt.Printf("Total duplicates found: %d\n", level6.Summary.Duplicates)
			if level6.Move != "" {
				fmt.Printf("Total items moved: %d\n", level6.Summary.Moves)
			} else if level6.Delete {
				fmt.Printf("Total items deleted: %d\n", level6.Summary.Deletes)
			}
			fmt.Printf("Began execution at: %s\n", level6.Summary.Start)
			fmt.Printf("Total execution time: %s\n", level6.Summary.Time)
		}
	}
}
