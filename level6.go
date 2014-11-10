package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
		if strings.Contains(path, "/.") || level6.MaxSize > 0 && f.Size <= level6.MaxSize {
			return err
		}
		if _, ok := level6.Files[f.Size]; !ok {
			level6.Files[f.Size] = make([]File, 0)
		}
		level6.Files[f.Size] = append(level6.Files[f.Size], f)
	}
	return err
}

func (level6 *Level6) GenerateHashes() {

	// prepare channel and wait group for asynchronous hashing
	sizes := make(chan int64, level6.MaxParallelism*2)
	var hashing sync.WaitGroup

	// prepare async bean counter
	hashes := make(chan int64, level6.MaxParallelism*2)
	var hashCount sync.WaitGroup

	// prepare go routines and add to wait group
	hashing.Add(level6.MaxParallelism)
	for i := 0; i < level6.MaxParallelism; i++ {
		go func() {
			defer hashing.Done()

			// use a single shared hash
			hash := sha256.New()

			// iterate each supplied size
			for size := range sizes {
				level6.Logger.Debug("channel %d, for file size %d", i, size)
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
					hashes <- 1
				}
			}
		}()
	}

	// count in parallel
	hashCount.Add(1)
	go func() {
		defer hashCount.Done()
		for _ = range hashes {
			level6.Summary.Hashes = level6.Summary.Hashes + 1
		}
	}()

	// send sets of files to each go routine
	for size, _ := range level6.Files {
		if len(level6.Files[size]) > 1 {
			sizes <- size
		}
	}

	// close channels and wait for done() calls before moving forward
	close(sizes)
	hashing.Wait()
	close(hashes)
	hashCount.Wait()
}

func (level6 *Level6) CompareHashes() {

	// prepare tasks and duplicates with a wait group
	sizes := make(chan int64, level6.MaxParallelism*2)
	var checking sync.WaitGroup

	// prepare bean counters for parallel duplicate appending
	duplicates := make(chan map[string][]File, level6.MaxParallelism*2)
	var counting sync.WaitGroup

	// prepare parallel duplicate checking
	checking.Add(level6.MaxParallelism)
	for i := 0; i < level6.MaxParallelism; i++ {
		go func() {
			defer checking.Done()
			for size := range sizes {
				dups := make(map[string][]File)
				for i, _ := range level6.Files[size] {
					if _, ok := dups[level6.Files[size][i].Hash]; !ok {
						dups[level6.Files[size][i].Hash] = make([]File, 0)
						for d := i + 1; d < len(level6.Files[size]); d++ {
							if level6.Files[size][i].Hash == level6.Files[size][d].Hash {
								dups[level6.Files[size][i].Hash] = append(dups[level6.Files[size][i].Hash], level6.Files[size][d])
							}
						}
						if len(dups[level6.Files[size][i].Hash]) > 0 {
							dups[level6.Files[size][i].Hash] = append(dups[level6.Files[size][i].Hash], level6.Files[size][i])
						}
					}
				}
				if len(dups) > 0 {
					duplicates <- dups
				}
			}
		}()
	}

	// capture and count duplicates in parallel (but in a single thread to remain safe)
	counting.Add(1)
	go func() {
		defer counting.Done()
		for dups := range duplicates {
			for hash, files := range dups {
				if len(files) > 0 {
					if _, ok := level6.Duplicates[hash]; !ok {
						level6.Duplicates[hash] = make([]File, 0)
					}
					level6.Duplicates[hash] = append(level6.Duplicates[hash], files...)
					level6.Summary.Duplicates = level6.Summary.Duplicates + 1
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

	// wait for all async operations to complete before closing down goroutines and moving forward
	close(sizes)
	checking.Wait()
	close(duplicates)
	counting.Wait()
}

func (level6 *Level6) Print() {
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

	// move or modify
	if level6.Move != "" {
		level6.MoveDuplicates()
	} else if level6.Delete {
		level6.DeleteDuplicates()
	}

	// summarize
	if level6.Summarize {
		if level6.Json {
			out, err := json.MarshalIndent(level6.Summary, "", "    ")
			if err == nil {
				fmt.Println(string(out))
			}
		} else {
			fmt.Println("Summary:")
			fmt.Printf("Total files scanned: %d\n", level6.Summary.Files)
			fmt.Printf("Total hashes generated: %d\n", level6.Summary.Hashes)
			fmt.Printf("Total duplicates found: %d\n", level6.Summary.Duplicates)
			if level6.Move != "" {
				fmt.Printf("Total items moved: %d\n", level6.Summary.Moves)
			} else if level6.Delete {
				fmt.Printf("Total items deleted: %d\n", level6.Summary.Deletes)
			}
			fmt.Printf("Total execution time: %s\n", time.Since(level6.Summary.Start))
		}
	}
}

func (level6 *Level6) MoveDuplicates() {
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
}

func (level6 *Level6) DeleteDuplicates() {
	for hash, _ := range level6.Duplicates {
		for i := 0; i < len(level6.Duplicates[hash])-1; i++ {
			err := os.Remove(level6.Duplicates[hash][i].Path)
			if err != nil {
				level6.Logger.Error("failed to delete file: %s, %s", level6.Duplicates[hash][i].Path, err)
			}
		}
	}
}
