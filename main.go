package main

import (
	// "fmt"
	"os"
	// "strings"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"
)

// file container
type FileBlob struct {
	Path string
	Size int64
	Hash string
}

// deduplication container
type Dedup struct {
	Logger log.Logger
	Path   string
	Delete bool
	Move   string
	Files  []FileBlob
}

func main() {

	// prepare new dedup struct /w logger
	d := Dedup{Logger: log.Logger{Level: log.INFO}}

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
	d.Path, _ = maps.String(&o, "", "path")
	d.Logger.Silent, _ = maps.Bool(&o, false, "quiet")
	d.Delete, _ = maps.Bool(&o, false, "delete")
	d.Move, _ = maps.String(&o, "", "move")
	if ok, _ := maps.Bool(&o, false, "verbose"); ok {
		d.Logger.Level = log.DEBUG
	}

	d.Logger.Debug("Dedup State: %+v", d)

	// steps to implement
	// - recurse directory
	// - build list of file paths
	// - store sizes with file paths
	// - sort files to group by size
	// - concurrently run through each group
	// - hash file contents and compare for each group

}

func (dedup *Dedup) Walk(path, f *os.FileInfo, err error) {

	// walk a supplied path and build a single map of file paths plus sizes
}

func (dedup *Dedup) Sort() {
	// sort files by size & return a map of grouped maps
}

func (dedup *Dedup) Process() {
	// spinup a queue of go routines to process each group individually
}
