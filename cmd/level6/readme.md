
# level6

This is the cli implementation and interface to the [level6](https://github.com/cdelorme/level6) library, allowing for simple file comparison to detect and deal with duplicate files.


## usage

Installation process:

    go get github.com/cdelorme/level6/...

_This will download the library and build all `cmd/` implementations._

Execution:

    level6 -t

_A non-zero exit code is produced when any errors are encountered._

For additional settings, try asking for help:

    level6 help

Finally we have embedded profiling when using the `GO_PROFILE` environment variable, which is treated as a file name.

An example with complete logging and profiling enabled:

	GO_PROFILE=profile LOG_LEVEL=debug ./level6 -t


## testing

Tests can be run via:

	LOG_LEVEL=silent go test -v -cover -race


## future

**The core implementation of the cli interface is effectively complete.**

At this point, the main goal is expanding and enhancing the library behind it.

Unless new functionality is added as "optional" behavior, no changes are expected to the interface.
