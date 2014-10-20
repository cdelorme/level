
# dev notes

Steps to process:

- recurse directory
- build list of file paths
- store sizes with file paths
- sort files to group by size
- concurrently run through each group
- hash file contents and compare for each group


## objectives

- cli
- gui

I'll start with a cli version, and create a GUI driven version in the future.


## cli options

- help (print usage)
- path (path to run against)
- quiet (no output & deletes)
- delete (permanently remove files)
- move (move duplicates to a supplied path instead)
- concurrency (number of concurrent processes to queue, what should the default be?)


## challenges

Had to build a utility to load cli flags, since the builtin was rather restrictive.

I suspect concurrency and memory management will be challenging, but that depends on how I implement some of the features.  I suspect at the very least the GC will not be very happy with me.

Implementing a queue for go routines to restrict the number of them that are spunup at a given time.  This should be fairly strait forward.

Lazy-hashing will be important for both cpu/disk io as well as memory consumption.  Also ensuring we only build the hash per file once, to avoid lots of extra overhead.
