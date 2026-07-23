package indexer

import (
	"context"
)

type Torrent struct {
	Title    string
	InfoHash string
	Magnet   string
	Size     uint64
	Tracker  string
	Files    uint

	TMDBTitle string
}

type SearchQuery struct {
	Type      string
	TMDBTitle string
	ID        string // TODO: we don't really need this
	Season    int
	Episode   int
}

type Indexer interface {
	Search(ctx context.Context, query SearchQuery) ([]Torrent, error)
	Name() string
}
