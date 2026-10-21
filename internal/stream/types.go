package stream

type SearchQuery struct {
	Type    string // Movie or TVShow
	ID      string
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

	TorrentTitle string
}

type StremioStream struct {
	Title        string   `json:"title,omitempty"`
	URL          string   `json:"url,omitempty"`
	InfoHash     string   `json:"infoHash,omitempty"`
	FileIdx      *int     `json:"fileIdx,omitempty"`
	Sources      []string `json:"sources,omitempty"`
	TorrentTitle string   `json:"torrentTitle,omitempty"`
}

type StreamResponse struct {
	Streams []StremioStream `json:"streams"`
}
