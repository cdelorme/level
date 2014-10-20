package main

import (
	// "fmt"
	// "os"
	// "strings"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"
)

// temporary settings wrapper
type Dedup struct {
	Path   string
	Quiet  bool
	Delete bool
	Move   string
}

func main() {

	// prepare new dedup struct
	d := Dedup{}

	// prepare logger
	logger := log.Logger{}

	// prepare cli options
	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("path", "path to begin scanning", "-p", "--path")
	appOptions.Flag("quiet", "silence output", "-q", "--quiet")
	appOptions.Flag("delete", "delete duplicate files", "-d", "--delete")
	appOptions.Flag("move", "move files to supplied path", "-m", "--move")
	// add a concurrency flag?
	o := appOptions.Parse()

	d.Path, _ = maps.String(&o, "", "path")
	d.Quiet, _ = maps.Bool(&o, false, "quiet")
	d.Delete, _ = maps.Bool(&o, false, "delete")
	d.Move, _ = maps.String(&o, "", "move")

	logger.Info("Dedup State: %+v", d)

	// steps to implement
	// - recurse directory
	// - build list of file paths
	// - store sizes with file paths
	// - sort files to group by size
	// - concurrently run through each group
	// - hash file contents and compare for each group

}

func (dedup *Dedup) Walk() {
	// walk a supplied path and build a single map of file paths plus sizes
}

func (dedup *Dedup) Sort() {
	// sort files by size & return a map of grouped maps
}

func (dedup *Dedup) Process() {
	// spinup a queue of go routines to process each group individually
}
