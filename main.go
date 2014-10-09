package main

import (
    "os"
    "fmt"
    "strings"
    "path"
    "os/user"

    "github.com/cdelorme/go-log"
    "github.com/cdelorme/go-config"
)

type Dedup struct {
    Delete bool
    Quiet bool
    Move string
}


func main() {

    // prepare new dedup struct
    d := Dedup{}

    // prepare logger
    logger := log.Logger{};

    // get user to acquire homedir
    u, e := user.Current()
    if e != nil {
        logger.Critical("No user?  Bugging out")
        os.Exit(1)
    }

    // attempt to load config data (my config library should totally assume default config file names by application)
    conf := config.Config{File: u.HomeDir + "/." + path.Base(os.Args[0])}
    conf.Load()

    // apply config or defaults
    // @todo (config lib missing bool option)

    // manually parsed flags
    for _, arg := range os.Args {
        if (arg == "-h" || arg == "--help") {
            // still trying to figure out how to override the default help info...
            fmt.Printf("%20s, %s\n", "-h | --help", "Print this usage information")
            fmt.Printf("%20s, %s\n", "-q | --quiet", "Quiet mode, no log output, will assume delete")
            fmt.Printf("%20s, %s\n", "-d | --delete", "Delete duplicate files")
            fmt.Printf("%20s, %s\n", "-m | --move", "Move duplicate files to supplied path")
            fmt.Printf("Example: %s\n", "")
        } else if arg == "-d" || arg == "--delete" {
            d.Delete = true
        } else if arg == "-q" || arg == "--quiet" {
            d.Quiet = true
        } else if arg == "-m" || strings.HasPrefix(arg, "--move") {
            fmt.Println("Moving Files")
            // need to add parsing logic here
        }
    }

    logger.Info("Test: %+v", d);
}
