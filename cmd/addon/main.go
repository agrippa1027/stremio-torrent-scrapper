package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"charm.land/log/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"scrapper/config"
	"scrapper/internal/jackett"
	"scrapper/internal/stream"
	"scrapper/internal/stremio"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"
	"scrapper/logging"
)

func main() {
	config, err := configLoad()
	if err != nil {
		log.Fatal("load config", "error", err)
		return
	}

	port := os.Getenv("PORT")

	// Passing down the logger to the addon
	// this seems reasonable as the creation of the logger is handle by very top level
	// and passed down to the addon
	logger := logging.Logger(os.Getenv("LOG_LEVEL"))

	httpClient := http.DefaultClient

	manifest := stremio.NewManifest(stremio.ManifestOptions{
		ID:          config.Manifest.ID,
		Version:     config.Manifest.Version,
		Name:        config.Manifest.Name,
		Description: config.Manifest.Description,
		Resources:   config.Manifest.Resources,
		Types:       config.Manifest.Types,
		Catalogs:    config.Manifest.Catalogs,
		Logo:        config.Manifest.Logo,
	})

	logger.Info("[MAIN] Manifest loaded", "manifest", manifest)

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Redis.Host + ":" + strconv.Itoa(config.Redis.Port),
		DB:   config.Redis.DB,
	})

	tmdbClient, err := tmdb.New(
		tmdb.Options{
			APIKey:         config.TMDB.APIKey,
			BaseURL:        config.TMDB.BaseURL,
			TimeoutSeconds: config.TMDB.TimeoutSeconds,
		},
		httpClient,
		redisClient,
		logger,
	)
	if err != nil {
		logger.Fatal("create tmdb client", "error", err)
		return
	}

	jackettClient, err := jackett.New(
		jackett.Options{
			BaseURL: config.Jackett.BaseURL,
			APIKey:  config.Jackett.APIKey,
		},
		httpClient,
	)
	if err != nil {
		logger.Fatal("create jackett client", "error", err)
		return
	}

	torboxClient, err := torbox.New(
		&torbox.Options{
			APIKey:    config.Torbox.APIKey,
			BaseURL:   config.Torbox.BaseURL,
			CacheOnly: config.Torbox.CachedOnly,
			Timeout:   config.Torbox.Timeout,
		},
		httpClient,
		logger,
	)
	if err != nil {
		logger.Fatal("create torbox client", "error", err)
		return
	}

	streamService, err := stream.New(
		tmdbClient,
		jackettClient,
		torboxClient,
		logger,
	)

	if err != nil {
		logger.Fatal("init stream service", "error", err)
		return
	}

	app, err := stremio.New(
		manifest,
		streamService,
		logger,
	)

	if err != nil {
		logger.Fatal("init addon", "error", err)
		return
	}

	mux := http.NewServeMux()
	app.Register(mux)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second}

	logger.Info("listening on port", "port", port)
	logger.Fatal("listen and serve", "err", srv.ListenAndServe())
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("load env from file", "error", err)
	}
}

func configLoad() (config.Config, error) {
	loadEnv()
	config, err := config.LoadConfigFile("config.yml")

	if err != nil {
		return config, err
	}

	config.Jackett.APIKey = os.Getenv("JACKETT_API_KEY")
	config.TMDB.APIKey = os.Getenv("TMDB_API_KEY")
	config.Torbox.APIKey = os.Getenv("TORBOX_API_KEY")

	return config, nil
}
