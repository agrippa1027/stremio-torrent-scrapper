package indexer

import "context"

type Result struct {
	Title    string
	InfoHash string
	Magnet   string
	Size     int64
	Seeds    int
	Source   string
	PageURL  string
}

type SearchQuery struct {
	Type    string
	ID      string
	Season  int
	Episode int
}

type Indexer interface {
	Search(ctx context.Context, query SearchQuery) ([]Result, error)
	Name() string
}
