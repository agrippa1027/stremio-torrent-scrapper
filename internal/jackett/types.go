package jackett

type QType string

const (
	QTypeMovie  QType = "movie"
	QTypeSeries QType = "series"
)

type Torrent struct {
	Title    string
	InfoHash string
	Magnet   string
	Size     uint64
	Tracker  string
	Files    uint
	Seeds    int
	Source   string

	TMDBTitle string
}

type SearchQuery struct {
	Type      QType
	TMDBTitle string
	ID        string
	Season    int
	Episode   int
}
