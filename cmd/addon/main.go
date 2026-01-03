package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"scrapper/internal/addon"

	"gopkg.in/yaml.v3"
)

func main() {
	cfg := addon.DefaultConfig()
	loadConfigFile("config.yml", &cfg)
	applyEnvOverrides(&cfg)

	port := cfg.Port

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	app, err := addon.New(cfg, logger)
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

func loadConfigFile(path string, cfg *addon.Config) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	_ = yaml.Unmarshal(data, cfg)
}

func applyEnvOverrides(cfg *addon.Config) {
	if v := os.Getenv("JACKETT_API_URL"); v != "" {
		cfg.JackettURL = v
	}
	if v := os.Getenv("JACKETT_API_KEY"); v != "" {
		cfg.JackettAPIKey = v
	}
	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}
}
