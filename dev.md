
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
- quiet (no output & deletes)
- delete (permanently remove files)
- move (move duplicates to a supplied path instead)


## challenges

Concurrency and Memory management may be challenging.

Some form of queue implementation to limit concurrent processes.

A similar queue mechanism to prevent storing more than a certain number of hashes when running comparisons.

