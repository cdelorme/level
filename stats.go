package level

const (
	StatsFiles      = "Total Files"
	StatsHashCrc32  = "Hashes Created (crc32)"
	StatsHashSha256 = "Hashes Created (sha256)"
	StatsDuplicates = "Duplicates Found"
	StatsDeleted    = "Moved Files"
	StatsMoved      = "Deleted Files"
)

type statCollector interface {
	Stats(stats)
}

type stats interface {
	Add(string, int)
}

type nilStats struct{}

func (n *nilStats) Add(_ string, _ int) {}
