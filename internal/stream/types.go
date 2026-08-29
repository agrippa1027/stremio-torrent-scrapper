package stream

type SearchQuery struct {
	Type    string // Movie or TVShow
	ID      string
	Title   string
	Season  int
	Episode int
}

type Stream struct {
	// Basic Info
	Title      string // normalized
	Year       string
	Seasons    []int // multiple for S01-S03
	Episodes   []int // multiple for S01E01E02; empty => pack
	Complete   bool
	Resolution string // "2160p", "1080p", ...
	Codec      string
	HDR        []string
	Uncensored bool
	Unrated    bool

	// Torrent Info
	InfoHash string
	Magnet   string
	Size     uint64
	Tracker  string
	Files    uint
}
