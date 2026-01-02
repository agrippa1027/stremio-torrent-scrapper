package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"scrapper/internal/addon"
)

func main() {
	cfg := addon.DefaultConfig()
	if v := os.Getenv("DATASET_PATH"); v != "" {
		cfg.DatasetPath = v
	}
	if v := os.Getenv("TRACKERS_PATH"); v != "" {
		cfg.TrackersPath = v
	}

	port := envOrDefault("PORT", "7000")

	app, err := addon.New(cfg)
	if err != nil {
		log.Fatalf("init addon: %v", err)
	}

	mux := http.NewServeMux()
	app.Register(mux)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
