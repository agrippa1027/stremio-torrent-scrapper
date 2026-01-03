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

type Indexer interface {
	Search(ctx context.Context, query string) ([]Result, error)
	Name() string
}
