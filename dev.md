
# dev notes

Revision of notes to accomodate changes to code and knowledge, paired with adjustments of objectives.

Code Steps:

- prepare `Level6` instance
- register switches
- parse switches & assign settings
- run dirwalk to build list of files
- run async hash generation
- sort by hashes
- run async deduplication

Originally I had planned on creating a gui wrapper right away, but I've decided to wait a bit to gain some knowledge of the state of go-sdl first.

In perhaps the further future, I would like to extend the comparison to do image, video, and audio comparisons using algorithms for those specific cases.


## cli switches

- help
- quiet
- verbose
- path
- delete
- move


## application behaviors

Switches:

- `quiet` overrides `verbose`
- `move` overrides `delete`
- `quiet` expects `delete` or `move`
- `path` is optional

The applications default behavior is to simply print out the duplicates to stdout.

It will exit if `quiet` is set, but neither `delete` nor `move` are assigned.

If no `path` has been supplied, it will use the current working directory.


### current blockers

I ran into some problems with how to efficiently handle building hashes and comparing them.  I have concluded that the solution is to perform these two actions entirely separately.  The complications necessary to build and compare hashes would likely eliminate any performance gains by not doing the comparison separately.

Implementing a queue was surprisingly easy, but not at all intuitive.

Setting up a two-way channel may prove to be a challenge, but a necessary one if I want to safely build a single map of duplicates by hash.


Logically, two files of different sizes cannot have the same contents and therefore we can safely ignore hash comparison of files with different sizes.  The code addresses this by building maps of arrays of files by size.  It only builds or compares hashes when there is more than one file in the group.

While we can safely assume two files of different sizes have different contents, the extremely low collision possibility exists, and to combat this we do not asynchronously group by hash when building a duplicates list.

I am not yet sure how well the application will work when run against a massive file set.  I may have to run some tests to see what peak memory consumption becomes when running the code against a massive set of files.


## tasks

- rename Dedup struct to Level6, and fix all references
- separately process file hashing and comparison
- build more efficient comparison
- add additional comparisons for image manipulation
- lazify implementation to check len of array of `File` objects before looping


# references

- [go-sdl2](https://github.com/veandco/go-sdl2)
