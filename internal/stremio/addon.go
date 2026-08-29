// Package stremio implements stremio addon logic.
// contains logic for handling stremio requests and interacting with external services.
package stremio

import (
	"net/http"
	"time"

	"scrapper/internal/jackett"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"

	"github.com/charmbracelet/log"

	"scrapper/internal/config"
)

type Addon struct {
	tmdbClient *tmdb.Client
	jackett    *jackett.Jackett
	manifest   Manifest
	logger     *log.Logger
	torbox     *torbox.Client
}

func New(cfg config.Config, logger *log.Logger) (*Addon, error) {
	httpClient := newHTTPClient()
	var err error

	tmdbClient, err := tmdb.New(cfg.TMDB, httpClient, logger)
	if err != nil {
		return nil, err
	}

	jackettClient, err := jackett.New(httpClient, logger)
	if err != nil {
		return nil, err
	}

	torboxClient, err := torbox.New(cfg.Torbox, httpClient, logger)
	if err != nil {
		return nil, err
	}

	manifest := NewManifest(cfg.Manifest)

	return &Addon{
		tmdbClient: tmdbClient,
		jackett:    jackettClient,
		manifest:   manifest,
		logger:     logger,
		torbox:     torboxClient,
	}, nil
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /manifest.json", a.manifestHandler)
	mux.HandleFunc("GET /stream/{type}/{id}", a.streamHandler)
}
