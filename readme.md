
# Linux File Deduplication
## Project name: level6_project

### Project Description

The objective of this software is file deduplication to reduce consumed space and clean out a path.

This is not a novel concept, but a learning experience and plan for enhancements going forward.


### Competing Software

An existing tool might be hardlink, which scans for duplicates and links the nodes to the same contents, safely accomplishing the objective.  A dry-run can simply print out duplicates.


### Project Details & Objectives

Personal objective is to learn about C and pointers.

Project objective is file deduplication command line tool with various advanced options, and potentially a Gtk3 addon.


#### Details List

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


### Notes

_Title inspired by To Aru Kagaku no Railgun's level6 project._

Myself, friends, and family all have issues with duplicate images, often not simply copies of the same data, but different modified versions of the same image.  Issues with cropped, rotated, discolored, or scaled instances of the same image change the hash significantly, making it impossible to deduplicate with a simple hash based system.  Ideally I wish to enhance the hash system with a secondary pass that performs advanced image detection for comparison of all the aforementioned issues.

This project is to outline one of my many personal ideas before accepting employer contracts that may prevent me from being able to embark on my ideas as my own and not as employer property.
