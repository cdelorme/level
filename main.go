package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"
)

// file container
type File struct {
	Path string
	Hash string `json:"-"`
}

type Level6 struct {
	MaxParallelism int
	Logger         log.Logger
	Path           string
	Delete         bool
	Move           string
	Files          map[int64][]File
	Duplicates     map[string][]File
}

// deduplication container
type Dedup struct {
	MaxParallelism int
	Logger         log.Logger
	Path           string
	Delete         bool
	Move           string
	Files          map[int64][]File
	Duplicates     map[string][]File
}

func main() {

	// prepare level6 /w logger and empty maps
	level6 := Level6{
		Logger:     log.Logger{Level: log.INFO},
		Files:      make(map[int64][]File),
		Duplicates: make(map[string][]File),
	}

	// optimize concurrent processing
	level6.MaxParallelism = runtime.NumCPU()
	runtime.GOMAXPROCS(level6.MaxParallelism)

	// get current directory
	cwd, _ := os.Getwd()

	// prepare cli options
	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("path", "path to begin scanning", "-p", "--path")
	appOptions.Flag("delete", "delete duplicate files", "-d", "--delete")
	appOptions.Flag("move", "move files to supplied path", "-m", "--move")
	appOptions.Flag("verbose", "verbose event output", "-v", "--verbose")
	appOptions.Flag("quiet", "silence all output", "-q", "--quiet")
	o := appOptions.Parse()

	// apply flags
	level6.Path, _ = maps.String(&o, cwd, "path")
	level6.Logger.Silent, _ = maps.Bool(&o, false, "quiet")
	level6.Delete, _ = maps.Bool(&o, false, "delete")
	level6.Move, _ = maps.String(&o, "", "move")
	if ok, _ := maps.Bool(&o, false, "verbose"); ok {
		level6.Logger.Level = log.DEBUG
	}

	// if quiet is set but not delete or move, exit
	if level6.Logger.Silent && !level6.Delete && level6.Move == "" {
		level6.Logger.Error("quiet is set but not delete or move, exiting...")
		return
	}

	// print initial level6 state
	level6.Logger.Debug("initial application state: %+v", level6)

	// build list of files grouped by size
	if err := filepath.Walk(level6.Path, level6.Walk); err != nil {
		level6.Logger.Error("failed to walk directory: %s", err)
	}
	level6.Logger.Debug("files: %+v", level6.Files)

	level6.GenerateHashes()
	level6.Logger.Debug("files /w hashes: %+v", level6.Files)

	// async compare
	level6.CompareHashes()
	level6.Logger.Debug("duplicates: %+v", level6.Duplicates)

	// print out contents as pretty json (readable, and programmatically parsable)
	if !level6.Logger.Silent {
		out, err := json.MarshalIndent(level6.Duplicates, "", "    ")
		if err == nil {
			fmt.Printf(string(out))
		}
	}

	if level6.Move != "" {
		// mkdir by hash & move files
		// handle name conflicts intelligently (prepend numbers pre-emptively)
	}

	if level6.Delete {
		// delete
	}

	// @todo
	// - separate structs and related methods to their own files
	// - write move & delete logic, and test on decent sized data sets (30~GB)
}

func (level6 *Level6) Walk(path string, file os.FileInfo, err error) error {
	if file.Mode().IsRegular() {
		size := file.Size()
		if _, ok := level6.Files[size]; !ok {
			level6.Files[size] = make([]File, 0)
		}
		level6.Files[size] = append(level6.Files[size], File{Path: path})
	}
	return err
}

func (level6 *Level6) GenerateHashes() {

	// prepare tasks channel and wait group for concurrent processing
	sizes := make(chan int64)
	var wg sync.WaitGroup

	// prepare go routines and add to wait group
	for i := 0; i < level6.MaxParallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// @todo test single hash instance
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
				}
			}
		}()
	}

	// send sets of files to each go routine
	for size, _ := range level6.Files {
		if len(level6.Files[size]) > 1 {
			sizes <- size
		}
	}

	// close channels and wait for done() calls before moving forward
	close(sizes)
	wg.Wait()
}

func (level6 *Level6) CompareHashes() {

	// prepare tasks and duplicates with a wait group
	sizes := make(chan int64)
	duplicates := make(chan map[string][]File)
	var wg sync.WaitGroup

	// prepare go routines and add to wait group
	for i := 0; i < level6.MaxParallelism; i++ {
		wg.Add(1)
		// go level6.CompareHashes(tasks, duplicates, &wg, i)
		go func() {
			defer wg.Done()
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

	// asynchronously capture duplicates but synchronously & safely append to level6.Duplicates
	go func() {
		for {
			dups := <-duplicates
			for hash, files := range dups {
				if len(files) > 0 {
					if _, ok := level6.Duplicates[hash]; !ok {
						level6.Duplicates[hash] = make([]File, 0)
					}
					level6.Duplicates[hash] = append(level6.Duplicates[hash], files...)
				}
			}
		}
	}()

	// send sizes to our channel and let our goroutines pick up the work
	for size, _ := range level6.Files {
		if len(level6.Files[size]) > 1 {
			sizes <- size
		}
	}

	// when all of our tasks are done, we can wait for completion then close our output
	close(sizes)
	wg.Wait()
	close(duplicates)
}
