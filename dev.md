
# Dev Notes

- Faster approach to group files by size first
- Then only compare hashes against same-size files

_This does not solve the image situation, but it would be faster than building hashes for all files right away._


### Goals

Would be ideal to have:

- An Library/API
- A Command Line
- A Gtk3 Gui


### High Level Ideas

- Run from Terminal, accepts flags and arguments.

Flags:

-h, --help - Displays usage information
-p, --path - Accepts a path to run from (default is /home)
-b, --binary - Runs complete binary check
-d, --delete - Accepts a path to move files to instead of deleting them

Examples:

This approach will scan the directory `/mnt/samba/backups` for duplicates by hash only and will permanently delete the files:

    dedup -p /mnt/samba/backups

This version will perform a full binary check on matching hash:

    dedup -b -p /mnt/samba/backups

This version will perform a full binary check and move the files to `/home/cdelorme/deleted` instead of deleting them:

    dedup -b -d /home/cdelorme/deleted -p /mnt/samba/backups


**Comprehensive details of operations:**

Checks flags and sets internal use variables (paths and binary check boolean).

Will create two temporary files in `/tmp` or `~/tmp`.  One stores the list of hashes and the full file paths.  The second will store a list of matches (Marked prefixes of "k" for kept, and "d" for deleted).

The program checks path or assumes `/home`.  It will perform an iterative directory walk.  Each file it encounters it will scan the binary and create an MD5 hash.

Once the directory walk is completed the full list is run through.  It takes the first item and compares the hash to every other item in the list.

When a match is found these steps occur:

- Check Binary Flag
    - On
        - Perform Full Binary Comparison using paths
            - Still Matched?
                - Yes
                    - Run Delete
                - No
                    - Move Forward
    - Off
        - Run Delete

Operation for Delete is as follows:

- Create a record in the second output file
- Delete the record from the first output file
- Is delete path set?
    - Yes
        - Move the file
    - No
        - Delete the file

When the run is completed it will delete the first record from the first file, and repeat it's operation from the main loop.


### Other

Investigating memory issues with regards to a large array of file names, sizes, and hashes.


