
# [level](https://github.com/cdelorme/level)

A cross platform FOSS library to scan for duplicate files written in [go](https://golang.org/).

**_Title inspired by the Japanese anime [Toaru Kagaku no Railgun](https://myanimelist.net/anime/6213/Toaru_Kagaku_no_Railgun)._**

**[Documentation](http://godoc.org/github.com/cdelorme/level)**


## design

All processing is done synchronously because the bottleneck will always be the persistent storage.

The primary operation (`LastOrder`) will scan a folder for files, discard all files with no size or which contain excluded segments, group the rest by size, then iterate the groups checking all but the first to ignore hard links using [`os.SameFile`](https://golang.org/pkg/os/#SameFile), read the remaining files two at a time in 4K chunks to [compare them byte-by-byte](https://golang.org/pkg/bytes/#Compare).

If run in test mode no further actions will take place, and it is expected that the caller will print the metrics collected and the groups of duplicates so the user may act upon them.

Otherwise, it will perform a weighted sort of each group favoring depth then frequency of directory discarding the first record with the lowest score so the rest may be deleted.

_If the file system uses a larger block size than the 4K buffer used by the software it **may** negatively affect the performance of the software._


## usage

Import the library:

    import "github.com/cdelorme/level"

_Please use the [godocs](http://godoc.org/github.com/cdelorme/level) for further instructions._

Installation process:

    go get github.com/cdelorme/level/...


## tests

Tests can be run via:

	go test -v -cover -race


## future

- add intelligent buffer size to detect disk block sizes and use the lowest common denominator


## concurrent

The addition of concurrency may improve performance so long as there is idle time spent processing data exceeding time spent reading files.

There are still race conditions with the implementation.
