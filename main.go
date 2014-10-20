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
	Delete bool
	Quiet  bool
	Move   string
}

func main() {

	// prepare new dedup struct
	d := Dedup{}

	// prepare logger
	logger := log.Logger{}

	// prepare cli options
	appOptions := option.App{Description: "file deduplication program"}
	appOptions.Flag("quiet", "silence output", "-q", "--quiet")
	appOptions.Flag("delete", "delete duplicate files", "-d", "--delete")
	appOptions.Flag("move", "move files to supplied path", "-m", "--move")
	o := appOptions.Parse()

	d.Quiet, _ = maps.Bool(&o, false, "quiet")
	d.Delete, _ = maps.Bool(&o, false, "delete")
	d.Move, _ = maps.String(&o, "", "move")

	logger.Info("Dedup State: %+v", d)

}
