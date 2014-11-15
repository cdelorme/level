package main

import "time"

type Summary struct {
	Files            int64
	Crc32Hashes      int64
	Crc32Duplicates  int64
	Sha256Hashes     int64
	Sha256Duplicates int64
	Moves            int64
	Deletes          int64
	Time             time.Time
}
