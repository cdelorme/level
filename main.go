package main

import (
	"crypto/sha256"
	"encoding/hex"
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
	Hash string
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

	// prepare new dedup struct /w logger, and empty files/dedup maps
	d := Dedup{
		Logger:     log.Logger{Level: log.INFO},
		Files:      make(map[int64][]File),
		Duplicates: make(map[string][]File),
	}

	// optimize concurrent processing
	d.MaxParallelism = runtime.NumCPU()
	runtime.GOMAXPROCS(d.MaxParallelism)

	// get current directory
	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	// prepare cli options
	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("path", "path to begin scanning", "-p", "--path")
	appOptions.Flag("delete", "delete duplicate files", "-d", "--delete")
	appOptions.Flag("move", "move files to supplied path", "-m", "--move")
	appOptions.Flag("verbose", "verbose event output", "-v", "--verbose")
	appOptions.Flag("quiet", "silence output", "-q", "--quiet")
	o := appOptions.Parse()

	// apply options to deduplication
	d.Path, _ = maps.String(&o, cwd, "path")
	d.Logger.Silent, _ = maps.Bool(&o, false, "quiet")
	d.Delete, _ = maps.Bool(&o, false, "delete")
	d.Move, _ = maps.String(&o, "", "move")
	if ok, _ := maps.Bool(&o, false, "verbose"); ok {
		d.Logger.Level = log.DEBUG
	}

	// print state of dedup before we continue
	d.Logger.Debug("initial application state: %+v", d)

	// test directory walk
	if err := filepath.Walk(d.Path, d.Walk); err != nil {
		d.Logger.Error("failed to walk directory: %s", err)
	}

	// print list of files grouped by size index
	d.Logger.Debug("files: %+v", d.Files)

	// run comparison process
	d.ComparisonProcess()

	// print duplicates
	d.Logger.Debug("duplicates: %+v", d.Duplicates)
	d.Logger.Debug("files: %+v", d.Files)

	// @todo
	// - fix broken hash assignment (not being attached to files)
	// - for optimal safety, run sort-by-hash logic after hash generation & move hash comparison to a second set of concurrent tasks
	// - fix and optimize broken deduplication comparison loops
	// - add methods to handle delete || move logic
	//     - if delete is set, move is ignored
	//     - move will generate folders with hash names and move files over, and must handle name conflicts (or simply number the files)
	//     - delete will remove all but the (first || last) indexed item
}

func (dedup *Dedup) Walk(path string, file os.FileInfo, err error) error {
	if !file.IsDir() {
		size := file.Size()
		if _, ok := dedup.Files[size]; !ok {
			dedup.Files[size] = make([]File, 0)
		}
		dedup.Files[size] = append(dedup.Files[size], File{Path: path})
	}
	return err
}

func (dedup *Dedup) ComparisonProcess() {

	// use a channel plus a waitgroup to manage processing in parallel
	tasks := make(chan int64)
	var wg sync.WaitGroup

	// prepare queue of workers
	for i := 0; i < dedup.MaxParallelism; i++ {
		wg.Add(1)
		go dedup.HashAndCompare(tasks, &wg, i)
	}

	// send a set of files to each goroutine
	for size, _ := range dedup.Files {
		tasks <- size
	}

	// close all channels, then wait for our waitgroup to finish
	close(tasks)
	wg.Wait()
}

func (dedup *Dedup) HashAndCompare(sizes chan int64, wg *sync.WaitGroup, num int) {
	defer wg.Done()

	// if we can share a single Hash class (one per channel) then we could improve this process

	// iterate each supplied size
	for size := range sizes {
		dedup.Logger.Debug("channel %d, size %d", num, size)

		for _, file := range dedup.Files[size] {
			content, err := ioutil.ReadFile(file.Path)
			if err != nil {
				dedup.Logger.Error("failed to parse file %s, %s", file.Path, err)
				continue
			}

			// create the hash
			hash := sha256.New()
			hash.Write(content)
			file.Hash = hex.EncodeToString(hash.Sum(nil))
			dedup.Logger.Debug("file %s hash %s", file.Path, file.Hash)
		}

		// run hash comparison to build a list of duplicates
		// for _, f := range dedup.Files[size] {
		// 	b := false
		// 	for _, d := range dedup.Files[size] {
		// 		if f != d && f.Hash == d.Hash {
		// 			if _, ok := dedup.Duplicates[f.Hash]; !ok {
		// 				dedup.Duplicates[f.Hash] = make([]File, 0)
		// 			}
		// 			dedup.Duplicates[f.Hash] = append(dedup.Duplicates[f.Hash], d)
		// 			b = true
		// 		}
		// 		if b {
		// 			dedup.Duplicates[f.Hash] = append(dedup.Duplicates[f.Hash], f)
		// 		}
		// 	}
		// }

	}
}
