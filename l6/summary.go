package level6

import "time"

type Summary struct {
	Files        int64
	Crc32Hashes  int64
	Sha256Hashes int64
	Duplicates   int64
	Moves        int64
	Deletes      int64
	Start        time.Time
	Time         time.Duration
}
