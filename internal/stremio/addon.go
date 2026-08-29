// Package stremio implements stremio addon logic.
// contains logic for handling stremio requests and interacting with external services.
package stremio

import (
	"net/http"
	"time"

	"scrapper/internal/jackett"
	"scrapper/internal/stream"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"

	"github.com/charmbracelet/log"
	"github.com/redis/go-redis/v9"

	"scrapper/internal/config"
)

type Addon struct {
	manifest      Manifest
	searchService *stream.StreamService
	tmdbClient    *tmdb.Client
	redisClient   *redis.Client
	logger        *log.Logger
}

func New(cfg config.Config, logger *log.Logger) (*Addon, error) {
	httpClient := newHTTPClient()
	var err error

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})

	if err != nil {
		return nil, err
	}

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

	searchService, err := stream.New(
		jackettClient,
		torboxClient,
		logger,
		time.Second*10,
	)

	if err != nil {
		return nil, err
	}

	return &Addon{
		manifest:      manifest,
		redisClient:   redisClient,
		tmdbClient:    tmdbClient,
		searchService: searchService,
		logger:        logger,
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
