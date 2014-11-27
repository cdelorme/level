
# level6

A cross platform, FOSS, file deduplication software for command line.

_Title inspired by To Aru Kagaku no Railgun and dark humor._


# alternatives

There is plenty of software that provides file deduplication, as well as modern file systems that have such features built-in.

However, almost no decent file deduplication utilities are both cross-platform and free.

A linux/unix builtin command `hardlink` may do a similar job of finding files with the same data, but it does not eliminate the duplicate content, rather it maps both file references to the same data, reducing consumed disk space but not the number of "files".


## sales pitch

As with anything I write, my aim is to provide the simplest implementation that works cleanly.

Reasons to use it:

- fast
- cross-platform
- free
- works from the command line
- finds duplicates and lists them
- can also move, or delete duplicates

Reasons to use it:

- it doesn't contain mounds of unit tests
- it doesn't carry heavy abstraction layers or complexity
- it's less than 600 lines of code

_While the program will build cross platform, Windows may have difficulty with the file walk, and also resource exhaustion forcing the OS to kill the program.  For windows please supply a feasible max size argument to alleviate resource consumption._


## application behavior

Switches:

- `path` is optional (uses pwd|cwd otherwise)
- `delete`
- `move` overrides `delete`
- `quiet` overrides `verbose` and expects `delete` or `move`
- `max-size` of files to hash in kilobytes
- `json` to print pretty or for sharing with another application
- `summary` to print summary data at end of execution
- `verbose` debug output
- `profile` produce a cpu profile file for `go test pprof`

The applications default behavior is to simply print out the duplicates to stdout.  Supplying the `json` flag will print the results as a json array.

It will exit if `quiet` is set, but neither `delete` nor `move` are assigned.

If no `path` has been supplied, it will use the current working directory.

It moves files after scanning and building hashes, which means your `move` folder can be inside `path`.

When `move` is selected, it will only move all but the first into the `move` path.

When `delete` is selected, it will remove all but the first identified instance.

To reduce load on the system you can specify a maximum size to build hashes, anything above that size will be ignored when hashing and comparing.

The `summary` option ignores `quiet`, and will print to json when `json` is set.

_Dot files are still included in the file total in the summary, but are not added to the files list nor is a hash generated or duplication processed on them._


### usage

To list duplicates in a given path:

    level6 -p /path/to/dedup/

To silently move files:

    level6 -q -m /path/to/temp/dir/

To silently delete files:

    level6 -q -d

For additional options:

    level6 -h


## building & installing

You can install with `go get github.com/cdelorme/level6`.

If you want to test it you can clone the repository and run `go get` and `go build`.  For convenience on osx or linux the `build` bash script will create a local gopath to test build in-place without affecting anything else on your system.


## hash comparison

To increase performnce, this project will generate crc32 hashes first, and only run sha256 against duplicates among the crc32 hashes.  This should greatly reduce processing time.

This project uses sha256 hashing to ensure unique file data, allowing you to make the decision to delete or move files with confidence.  To further reduce possible conflicts, it only compares hashes of files of equal size.

[![sha256 is preferred](http://i.stack.imgur.com/46Vwb.jpg)](http://crypto.stackexchange.com/questions/1170/best-way-to-reduce-chance-of-hash-collisions-multiple-hashes-or-larger-hash)


## future plans

- optional high-fidelity flag for full-binary comparison
    - optionally performed on any sha256 matches
    - can also set max-size to a sane default, and run full-binary for files above that size limit
- implement complex image, video, and audio comparison algorithms with support for things like
    - partial image, video, or audio clips (including images from video)
    - images that have been:
        - scaled
        - rotated
        - cropped
        - discolored
- create a gui interface wrapper for the cli

_Some ideas for comparison include keypoint detection and decision trees, but I'll have to do a bit of research before I'm ready to take the code to the next level._
