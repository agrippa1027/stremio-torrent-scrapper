// Package addon implements the main logic for the media streaming addon,
// including handling requests, managing indexers, and interacting with external services like Jackett and Torbox.
package addon

import (
	"net/http"

	"scrapper/internal/addon/stremio"
	"scrapper/internal/indexer"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"

	"scrapper/internal/addon/config"

	"github.com/charmbracelet/log"
)

type Addon struct {
	indexers         []indexer.Indexer
	manifest         stremio.Manifest
	logger           *log.Logger
	torbox           *torbox.Client
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

	return &Addon{
		indexers:         indexers,
		manifest:         cfg.Manifest,
		logger:           logger,
		torbox:           torboxClient,
	}, nil
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /manifest.json", a.manifestHandler)
	mux.HandleFunc("GET /stream/{type}/{id}", a.streamHandler)
}
