
# [level](https://github.com/cdelorme/level)

A cross platform, FOSS, duplicate file detection and management software library written in the [go programming language](https://golang.org/).

**_Title inspired by the Japanese anime "Toaru Kagaku no Railgun"._**


## design

This project has been updated to run an entire pipeline of operations including support for user-defined custom steps.  A simple, well-defined interface is exposed and supported from the core utility.

We run all steps in a loop, getting a fresh list of files and acquiring a list of duplicates to move or delete.

All errors are captured and supplied from the core `Execute()`.


### step

Each part of the pipeline is a `step`, which is run in a loop that requires a `filepath.WalkFunc` and `Execute() ([][]string, error)` function, and optionally supports `Logger(logger)` and `Stats(stats)` to share core resources.

By design it will run the move/delete per step and rerun a fresh walk.  **This reduces redundancy when checking files that would have been removed by earlier steps in the pipeline.**

The core walk calls the step walk, and pre-emptively filters excluded files by path-match, including dot-files (eg. `/.`) the move path, and any other values supplied through the core interface `Excludes` property parsed as a comma-delimited array.

To safely deal with moving files without creating duplicates we try `os.Rename` then `io.Copy`, opening the input file with Read/Write to avoid copying read-only files that we cannot delete.


#### data

Currently the only completed internal step is a fulld ata comparison.

It groups files by size, and runs crc32, sha256, `os.SameFile`, and finally byte comparison against files in each group.

[![sha256 is preferred](http://i.stack.imgur.com/46Vwb.jpg)](http://crypto.stackexchange.com/questions/1170/best-way-to-reduce-chance-of-hash-collisions-multiple-hashes-or-larger-hash)

_In truth, a sha256 should never have duplicates that don't lead to a byte comparison match, but this is just an extra step for safety._

The `os.SameFile` helps us avoid deleting hardlinks and symlinks.

**Currently this process is synchronous which makes sense as it is bound by diskio, _however future plans include high levels of concurrency_.**

All operations are buffered and only use 32kb per buffer, which helps us evade resource exhaustion.


## usage

Import the library:

    import "github.com/cdelorme/level"

Installation process:

    go get github.com/cdelorme/level/...

_This will download the library and build all `cmd/` implementations._


## problems

Currently the core library lacks unit tests to validate behavior.

In spite of the disk io being the bottleneck, with buffered operations we can still gain significant performance with proper groupings of goroutines.  _Working on a queueing mechanism that can be shared across many steps of the pipeline._

Currently the move/delete behavior ignores groups of same-path duplicates, which leads to staggered removal of duplicates.  _Ideally we would like to move whole groups of duplicates by path, and cleanup empty folders afterwards._


## future

- add comprehensive unit tests
- add logic for duplicates to be grouped by path so we can correctly delete matching sets
- add logic to check post-remove for an empty parent folder to cleanup after our process
- add a well designed queue mechanism for standardized concurrent processing
- add filetype/mimetype detection to `video.go`, `image.go`, `audio.go`, and `content.go`
- add content comparison of test files using size as a measure prior to scanning
	- add behavior to log partial matches beyond a certain percent
- add comparison of video/images, and audio (scaled, cropped, rotated, discolored, distorted, in-motion, timelapsed, etc...)
	- if possible would prefer native support in go for complete cross platform capability
	- video step should simply extract frames as images to run image comparisons
	- advanced keypoint detection and decision-trees for image comparison

_If we use 2000 goroutines from our queue with two buffered files at a time we can expect around 120mb of diskio at any given time, which is still reasonably small._
