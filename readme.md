
# level6

A FOSS file deduplication software for command line.

_Title inspired by To Aru Kagaku no Railgun and dark humor._


# alternatives

There is plenty of software that provides file deduplication, as well as modern file systems that have such features built-in.  However, almost no decent file deduplication tools are both cross-platform and free.

A linux/unix builtin command `hardlink` may do a similar job of finding files with the same data, but it does not eliminate the duplicate content, rather it maps both file references to the same data, reducing consumed disk space but not the number of files on record.


## sales pitch

working on it...


## usage

tbdt


## building & installing

On linux or osx you may run the `build` bash script.  It will create a project-local `GOPATH` for compiling the software without affecting the rest of your system.  Otherwise the standard `go get`, `go build`, and `go install` commands will work.

The program requires no configuration files and can be installed directly to a `bin` folder.  If you have your own `$GOPATH/bin` in `$PATH`, then you can simply `go get github.com/cdelorme/level6`.


## future plans

1. Accept arguments:

-q --quiet, no output to terminal
-l --log, file to log output (if -l and not -q print AND write to file)
-h --hashtype, specify md5 or sha, or others (if able) for the process (allows variant degrees of reliability)
-d --delete, automatically delete duplicates
-m --move, automatically relocate duplicates

2. Usign root OR a supplied path build an index of file information, including hash (md5 or sha1), full path, etc...

3. Sort the list by hash, and build a list of duplicates to display and/or log

4. Create a gtk3 interface to overlay the command with visual options

5. Enhance with new options for image detection, accounting for:

- scaling
- rotation
- cropping
- discoloration

_Some ideas for this include keypoint detection and decision trees, but more research needs to be done before this can be looked at further._
