
# level6

A cross platform, FOSS, file deduplication software library.

**_Title inspired by the Japanese anime "Toaru Kagaku no Railgun"._**


## design

This project is designed to scan files in a supplied folder, group by size, then apply layers of hashing to identify duplicates efficiently and in parallel.  It automatically excludes files and folders that begin with a period (_including files within folders that begin with a period_), as these are considered "hidden files".

[![sha256 is preferred](http://i.stack.imgur.com/46Vwb.jpg)](http://crypto.stackexchange.com/questions/1170/best-way-to-reduce-chance-of-hash-collisions-multiple-hashes-or-larger-hash)

The hashing is optimized by running crc32 before sha256 to reduce the number of sha256 hashes needed.  All hashing is buffered reducing the footprint to 32kb per parallel process according to the `io.Copy()` default buffer size.  _This buffer should alleviate problems with resource exhaustion when run on a Windows machine._

Once duplicates have been assembled the settings are used to either print the results, move, or delete files.  In the process of moving files we attempt to create the parent directories, and we try `os.Rename` followed by `io.Copy` as a fallback (necessary if moving files to another disk).  _Currently this leads to problems if the files are read-only, where we may create a **new duplicate** in the supplied move path and be unable to delete the original._

A summary will be printed including execution time, and stats collected during execution.

All errors will be logged, but only the latest error will be returned to the executor.


## problems

This project is still an early-draft and was originally written as a proof-of-concept leveraging channel-based-concurrency.

The original did not use buffers and loaded whole files into memory, which became a problem when dealing with very large files such as VM images, and would get killed by Windows resource manager due to "Resource Exhaustion".

_While this has been addressed with buffers, it may still be incorrect to leverage parallel processing against the disk which is a slower component._  I plan to optimize and benchmark to find a happy balance.

The code is subject to race-conditions as no protection is used when dealing with maps.  In particular the code is written to write to maps synchronously, but read in parallel.  This is a dangerous game, _but was a decision to optimize to avoid more complexity with passing large amounts of data around or adding many more channels._

There are other "bugs" in the code, specifically the lack of duplicate grouping and cleanup.  This means if two folders exist the item deleted or moved may come from a mixture of the two, and if any folders are left empty we do not cleanup the folder afterwards.  _These I intend to address in future iterations._

Finally because only one function is exposed its purpose as a "Library" is rather limited in scope (which was not the intention).


## usage

Import the library:

    import "github.com/cdelorme/level6"

Installation process:

    go get github.com/cdelorme/level6/...

_This will download the library and build all `cmd/` implementations._


## future

- correctly detect read-only permissions and fail on `move()`
- create package-level global functions for common library functionality
- byte-by-byte comparison as final data-check after sha256
- add logic for duplicates to be grouped by path so we can correctly delete matching sets
- add logic to check post-remove for an empty parent folder to cleanup after our process
- add comprehensive unit tests
- optimize concurrent processing and management
- add file-type and categorization component
    - to capture text, audio, video, and image files
- add content comparison for text files
    - _partial matches can be printed but not augmented_
- add visual comparison functions
    - _should be compatible with video files_
    - go-native support for running scaled, cropped, rotated, discolored, in-motion comparisons would be ideal
- add audible comparison functions
    - _this is a completely new field for me, so I have no clue where to start_

_At a low level I understand that keypoint detection and decision-trees are used in image comparison, but afaik none of these exist yet as native go features._

Choosing sane defaults and adding flags to enable, disable, and augment behavior when doing the more touchy duplicate-check logic would be good to consider at the cli level.
