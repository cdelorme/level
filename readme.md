
# level6

A FOSS file deduplication software for command line.

_Title inspired by To Aru Kagaku no Railgun and dark humor._

## sales pitch

There are plenty of softwares that provide file deduplication, as well as modern file systems that have such features built-in.

However, almost no decent file deduplication tools are both cross-platform and free.

A similar cli tool might be `hardlink`.


## usage

tbdt


## building

Simply run `go get` then `go install` to download any dependencies and then install the command to your golang bin local to your system.

Alternatively on linux or mac systems you may use the `dev` script to set a project-local `GOPATH` and then run `go get` and `go build` to create the executable in the project folder and test without effecting your local `GOPATH` contents.


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


# references

- []()

