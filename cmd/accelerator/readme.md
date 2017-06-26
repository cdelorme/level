
# [accelerator](https://github.com/cdelorme/level/tree/master/cmd/accelerator)

This is the command line implementation and interface to the [level](https://github.com/cdelorme/level) library, which finds and deletes all but one of the duplicates and print the metrics collected.

If run in test mode it will print the groups of duplicates instead of deleting all but one of them.

Any non-fatal errors encountered will be logged, but will neither halt execution nor be reflected by the exit code.


## usage

Installation process:

    go get github.com/cdelorme/level/...

_This will download the library and build all `cmd/` implementations._

Execution (_with test mode enabled_):

    accelerator -t

A summary is tracked and printed from the cli implementation, including the start time, and stats collected by steps in the pipeline.

_A non-zero exit code is produced when any errors are encountered._

For additional settings, try asking for help:

    accelerator help
