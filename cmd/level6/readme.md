
# level6

This is the cli implementation and interface to the [level6](https://github.com/cdelorme/level6) library, allowing for simple file comparison to detect and deal with duplicate files.


## usage

Installation process:

    go get github.com/cdelorme/level6/...

_This will download the library and build all `cmd/` implementations._

Execution (_with test mode enabled_):

    level6 -t

A summary is tracked and printed from the cli implementation, including the start time, and stats collected by steps in the pipeline.

_A non-zero exit code is produced when any errors are encountered._

For additional settings, try asking for help:

    level6 help

Finally we have embedded profiling when using the `GO_PROFILE` environment variable, which is treated as a file name.

An example with complete logging and profiling enabled:

	GO_PROFILE=profile LOG_LEVEL=debug ./level6 -t


## testing

Tests can be run via:

	LOG_LEVEL=silent go test -v -cover -race

_The `LOG_LEVEL` setting omits a forced crash from test execution._


## future

**The core implementation of the cli interface is complete.**

At this point, the main goal is expanding and enhancing the library behind it and making enhancements in behavior transparent to its usage.

_The exception is if we add "optional" functionality and need more flags to enable or disable those features._
