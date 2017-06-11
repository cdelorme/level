
# [level](https://github.com/cdelorme/level)

A cross platform FOSS library to scan for duplicate files written in [go](https://golang.org/).

**_Title inspired by the Japanese anime "Toaru Kagaku no Railgun"._**

**[Documentation](http://godoc.org/github.com/cdelorme/level)**


## design

Initially designed using a pipeline with parallel processing, has undergone multiple changes to address performance, resource consumption, and code simplicity.

All operations are publicly accessible and independently operable.

The primary function of the library is to scan a folder for files, group them by size, and compare them using progressively more expensive operations to rule out non-duplicates quickly.  **The current operations include crc32, sha256, and finally byte-by-byte comparison for 100% accuracy.**

[![sha256 is preferred](http://i.stack.imgur.com/46Vwb.jpg)](http://crypto.stackexchange.com/questions/1170/best-way-to-reduce-chance-of-hash-collisions-multiple-hashes-or-larger-hash)

While the likelihood of a sha256 collision is astronomically small, the byte-by-byte comparison is written as a final checkpoint and with the assumption that it is possible.

Any hard links will be detected and ignored using [`os.SameFile`](https://golang.org/pkg/os/#SameFile).

_Since the scan assumes files will have the same byte-size, if a scan is run across multiple hard disks with varying block sizes then it may not accurately detect duplicates between the disks._

As the primary bottleneck will always be disk io, all processing is synchronous.


## usage

Import the library:

    import "github.com/cdelorme/level"

_Please use the [godocs](http://godoc.org/github.com/cdelorme/level) for further instructions._

Installation process:

    go get github.com/cdelorme/level/...


## notes

Some consideration has been given to producing a web-interface for this tool, however the code was not written in a way that can pause and resume operations.  _This change would be a significant undertaking._

All alternative forms of deduplication are below 100% accuracy; hence content, image, audio, and video comparison have been scrapped as part of this project.  _Also the difficulty of writing the necessary code to read the file formats and the algorithms to compare the data extend far beyond what I am currently capable of._

Switching to buffered hashing and byte comparison eliminates resource exhaustion.  _It may be possible to further enhance this by checking the hard disk on the machine for optimal buffer size, then using [`io.CopyBuffer`](https://golang.org/pkg/io/#CopyBuffer) to use a specific size._


# references

- [buffered sha256](http://stackoverflow.com/questions/15879136/how-to-calculate-sha256-file-checksum-in-go)
- [weighted byte comparison](https://golang.org/pkg/bytes/#Compare)
- [buffered line comparison](https://play.golang.org/p/NlQZRrW1dT)
