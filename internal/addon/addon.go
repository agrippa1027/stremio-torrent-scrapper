package addon

import (
	"fmt"
	"log/slog"
	"net/http"

	"scrapper/internal/indexer"
)

type Addon struct {
	indexers []indexer.Indexer
	trackers []string
	manifest Manifest
	logger   *slog.Logger
}

func New(cfg Config, logger *slog.Logger) (*Addon, error) {
	if cfg.JackettURL == "" || cfg.JackettAPIKey == "" {
		return nil, fmt.Errorf("missing Jackett config: set JACKETT_API_URL and JACKETT_API_KEY")
	}

	trackers := DefaultTrackers()
	client := indexer.NewHTTPClient()
	jackettIndexer, err := indexer.NewJackett(client, cfg.JackettURL, cfg.JackettAPIKey)
	if err != nil {
		return nil, err
	}
	indexers := []indexer.Indexer{jackettIndexer}

	if logger == nil {
		logger = slog.Default()
	}

	return &Addon{
		indexers: indexers,
		trackers: trackers,
		manifest: cfg.Manifest,
		logger:   logger,
	}, nil
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("/manifest.json", a.manifestHandler)
	mux.HandleFunc("/stream/", a.streamHandler)
}
