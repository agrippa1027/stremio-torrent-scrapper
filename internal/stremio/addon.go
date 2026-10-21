// Package stremio implements stremio addon logic.
// contains logic for handling stremio requests and interacting with external services.
package stremio

import (
	"fmt"
	"net/http"

	"scrapper/internal/stream"

	"github.com/charmbracelet/log"
)

type Addon struct {
	manifest      Manifest
	searchService *stream.StreamService
	logger        *log.Logger
}

func New(
	manifest Manifest,
	searchService *stream.StreamService,
	logger *log.Logger,
) (*Addon, error) {
	var err error

	if err != nil {
		return nil, fmt.Errorf("failed to create search service: %w", err)
	}

	return &Addon{
		manifest:      manifest,
		searchService: searchService,
		logger:        logger,
	}, nil
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /manifest.json", a.manifestHandler)
	mux.HandleFunc("GET /stream/{type}/{id}", a.streamHandler)
}
