// Package addon implements the main logic for the media streaming addon,
// including handling requests, managing indexers, and interacting with external services like Jackett and Torbox.
package addon

import (
	"fmt"
	"net/http"

	"scrapper/internal/indexer"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"

	"github.com/charmbracelet/log"
)

type Addon struct {
	indexers         []indexer.Indexer
	manifest         Manifest
	logger           *log.Logger
	torbox           *torbox.Client
	torboxCachedOnly bool
}

func New(cfg Config, logger *log.Logger) (*Addon, error) {
	if cfg.JackettURL == "" || cfg.JackettAPIKey == "" {
		return nil, fmt.Errorf("missing Jackett config: set JACKETT_API_URL and JACKETT_API_KEY")
	}
	if cfg.TorboxAPIKey == "" {
		return nil, fmt.Errorf("missing Torbox config: set TORBOX_API_KEY")
	}

	client := indexer.NewHTTPClient()
	var err error

	var tmdbClient *tmdb.Client
	if cfg.TMDBAPIKey != "" {
		tmdbClient, err = tmdb.NewClient(cfg.TMDBAPIKey, client)
		if err != nil {
			return nil, err
		}
	}

	jackettIndexer, err := indexer.NewJackett(client, cfg.JackettURL, cfg.JackettAPIKey, tmdbClient, logger)
	if err != nil {
		return nil, err
	}
	indexers := []indexer.Indexer{jackettIndexer}
	torboxClient, err := torbox.NewClient(cfg.TorboxBaseURL, cfg.TorboxAPIKey, client)
	if err != nil {
		return nil, err
	}

	return &Addon{
		indexers:         indexers,
		manifest:         cfg.Manifest,
		logger:           logger,
		torbox:           torboxClient,
		torboxCachedOnly: cfg.TorboxCachedOnly,
	}, nil
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /manifest.json", a.manifestHandler)
	mux.HandleFunc("GET /stream/{type}/{id}", a.streamHandler)
	// NOTE: what is the point of this
	// this should be not required
	mux.HandleFunc("GET /torbox/stream/{hash}", a.torboxStreamHandler)
}
