package main

import "time"

type Summary struct {
	Files      int64
	Duplicates int64
	Hashes     int64
	Moves      int64
	Deletes    int64
	Start      time.Time
}
