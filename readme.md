
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
- it's less than 300 lines of code

_While I would like to provide cross-platform compatibility, currently there are some issues getting the code to work in Windows, specifically the memory management and `FileInfo` seem to have bugs either in the windows go implementation or a possible memory leak._


## application behavior

Switches:

- `quiet` overrides `verbose`
- `move` overrides `delete`
- `quiet` expects `delete` or `move`
- `path` is optional
- `json` to print pretty or for sharing with another application

The applications default behavior is to simply print out the duplicates to stdout.  Supplying the `json` flag will print the results as a json array.

It will exit if `quiet` is set, but neither `delete` nor `move` are assigned.

If no `path` has been supplied, it will use the current working directory.

It moves files after scanning and building hashes, which means your `move` folder can be inside `path`.

When `move` is selected, it will only move all but the first into the `move` path.

When `delete` is selected, it will remove all but the first identified instance.


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

This project uses sha256 hashing to ensure unique file data.  It also only compares two files with equal sizes.

[![sha256 is preferred](http://i.stack.imgur.com/46Vwb.jpg)](http://crypto.stackexchange.com/questions/1170/best-way-to-reduce-chance-of-hash-collisions-multiple-hashes-or-larger-hash)


## future plans

- implement complex image, video, and audio comparison algorithms with support for things like
    - partial image, video, or audio clips (including images from video)
    - images that have been:
        - scaled
        - rotated
        - cropped
        - discolored
- create a gui interface wrapper for the cli

_Some ideas for comparison include keypoint detection and decision trees, but I'll have to do a bit of research before I'm ready to take the code to the next level._
