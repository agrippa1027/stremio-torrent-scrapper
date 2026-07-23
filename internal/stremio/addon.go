// Package stremio implements stremio addon logic.
// contains logic for handling stremio requests and interacting with external services.
package stremio

import (
	"net/http"

	"scrapper/internal/indexer"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"

	"github.com/charmbracelet/log"

	"scrapper/internal/config"
)

type Addon struct {
	indexers []indexer.Indexer
	manifest Manifest
	logger   *log.Logger
	torbox   *torbox.Client
}

func New(cfg config.Config, logger *log.Logger) (*Addon, error) {
	client := indexer.NewHTTPClient()
	var err error

	tmdbClient, err := tmdb.NewClient(cfg.TMDB, client)
	if err != nil {
		return nil, err
	}

	jackettIndexer, err := indexer.NewJackett(client, tmdbClient, logger)
	if err != nil {
		return nil, err
	}

	indexers := []indexer.Indexer{jackettIndexer}

	torboxClient, err := torbox.NewClient(cfg.Torbox, client)
	if err != nil {
		return nil, err
	}

	manifest := NewManifest(cfg.Manifest)

	return &Addon{
		indexers: indexers,
		manifest: manifest,
		logger:   logger,
		torbox:   torboxClient,
	}, nil
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /manifest.json", a.manifestHandler)
	mux.HandleFunc("GET /stream/{type}/{id}", a.streamHandler)
}
