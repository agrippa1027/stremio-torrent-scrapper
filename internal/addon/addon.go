package addon

import (
	"net/http"

	"scrapper/internal/indexer"
)

type Addon struct {
	indexers []indexer.Indexer
	trackers []string
	manifest Manifest
}

func New(cfg Config) (*Addon, error) {
	trackers := DefaultTrackers()
	client := indexer.NewHTTPClient()
	indexers := []indexer.Indexer{
		indexer.NewYTS(client),
		indexer.NewX1337(client),
	}

	return &Addon{
		indexers: indexers,
		trackers: trackers,
		manifest: cfg.Manifest,
	}, nil
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("/manifest.json", a.manifestHandler)
	mux.HandleFunc("/stream/", a.streamHandler)
}
