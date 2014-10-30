package main

import (
	// "fmt"
	"os"
	"path/filepath"
	// "strings"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"
)

// file container
type FileBlob struct {
	Path string
	Hash string
}

// deduplication container
type Dedup struct {
	Logger log.Logger
	Path   string
	Delete bool
	Move   string
	Files  map[int64][]FileBlob
}

func main() {

	// prepare new dedup struct /w logger & empty file map
	d := Dedup{Logger: log.Logger{Level: log.INFO}, Files: make(map[int64][]FileBlob)}

	// get current directory
	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	// prepare cli options
	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("path", "path to begin scanning", "-p", "--path")
	appOptions.Flag("delete", "delete duplicate files", "-d", "--delete")
	appOptions.Flag("move", "move files to supplied path", "-m", "--move")
	appOptions.Flag("verbose", "verbose event output", "-v", "--verbose")
	appOptions.Flag("quiet", "silence output", "-q", "--quiet")
	// add a concurrency flag?
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

	// steps to implement
	// - concurrently run through each group
	// - hash file contents and compare for each group

}

func (dedup *Dedup) Walk(path string, file os.FileInfo, err error) error {
	if !file.IsDir() {
		size := file.Size()
		if _, ok := dedup.Files[size]; !ok {
			dedup.Files[size] = make([]FileBlob, 0)
		}
		dedup.Files[size] = append(dedup.Files[size], FileBlob{Path: path})
	}
	return err
}

func (dedup *Dedup) Sort() {
	// sort files by size & return a map of grouped maps
}

func (dedup *Dedup) Process() {
	// spinup a queue of go routines to process each group individually
}
