
# [level](https://github.com/cdelorme/level)

A cross platform FOSS library to scan for duplicate files written in [go](https://golang.org/).

**_Title inspired by the Japanese anime "Toaru Kagaku no Railgun"._**

**[Documentation](http://godoc.org/github.com/cdelorme/level)**


## design

All processing is done synchronously because the bottleneck will always be the persistent storage.

The primary function of the library is to scan a folder for files, discard files containing excluded segments, group them by size, discard hard-links by checking [`os.SameFile`](https://golang.org/pkg/os/#SameFile), and [compare them byte-by-byte](https://golang.org/pkg/bytes/#Compare).

_Since the scan assumes files will have the same byte-size, if a scan is run across multiple hard disks with varying block sizes then it may not accurately detect duplicates between the disks._


## usage

Import the library:

    import "github.com/cdelorme/level"

_Please use the [godocs](http://godoc.org/github.com/cdelorme/level) for further instructions._

Installation process:

    go get github.com/cdelorme/level/...


## tests

Tests can be run via:

	go test -v -cover -race
