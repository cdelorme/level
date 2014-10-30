package main

import (
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
}

func main() {

	// prepare new dedup struct /w logger & empty file map
	d := Dedup{Logger: log.Logger{Level: log.INFO}, Files: make(map[int64][]File)}

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
	d.Logger.Debug("Dedup State: %+v", d)

	// test directory walk
	if err := filepath.Walk(d.Path, d.Walk); err != nil {
		d.Logger.Error("Failed to walk directory: %s", err)
	}

	// print list of files grouped by size index
	d.Logger.Debug("Files: %+v", d.Files)

	// run comparison process
	d.ComparisonProcess()

	d.Logger.Critical("Testing %d and %d", 1, 4)

	// steps to implement
	// - concurrently run through each group
	// - hash file contents and compare for each group

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

	// iterate each supplied size
	for size := range sizes {
		dedup.Logger.Debug("Channel %d, size %d", num, size)
	}
}
