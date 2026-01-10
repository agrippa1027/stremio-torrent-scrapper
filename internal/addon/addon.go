package addon

import (
	"fmt"
	"log/slog"
	"net/http"

	"scrapper/internal/indexer"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"
)

type Addon struct {
	indexers         []indexer.Indexer
	trackers         []string
	manifest         Manifest
	logger           *slog.Logger
	torbox           *torbox.Client
	torboxCachedOnly bool
}

func New(cfg Config, logger *slog.Logger) (*Addon, error) {
	if cfg.JackettURL == "" || cfg.JackettAPIKey == "" {
		return nil, fmt.Errorf("missing Jackett config: set JACKETT_API_URL and JACKETT_API_KEY")
	}
	if cfg.TorboxAPIKey == "" {
		return nil, fmt.Errorf("missing Torbox config: set TORBOX_API_KEY")
	}

	trackers := DefaultTrackers()
	client := indexer.NewHTTPClient()
	var tmdbClient *tmdb.Client
	var err error
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

	if logger == nil {
		logger = slog.Default()
	}

	return &Addon{
		indexers:         indexers,
		trackers:         trackers,
		manifest:         cfg.Manifest,
		logger:           logger,
		torbox:           torboxClient,
		torboxCachedOnly: cfg.TorboxCachedOnly,
	}, nil
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("/manifest.json", a.manifestHandler)
	mux.HandleFunc("/stream/", a.streamHandler)
	mux.HandleFunc("/torbox/stream/", a.torboxStreamHandler)
}
