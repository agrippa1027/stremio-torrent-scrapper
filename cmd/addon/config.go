package main

import (
	"os"
	"strings"

	"scrapper/internal/addon"

	"gopkg.in/yaml.v3"
)

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
	if v := os.Getenv("TMDB_API_KEY"); v != "" {
		cfg.TMDBAPIKey = v
	}
	if v := os.Getenv("TORBOX_API_KEY"); v != "" {
		cfg.TorboxAPIKey = v
	}
	if v := os.Getenv("TORBOX_BASE_URL"); v != "" {
		cfg.TorboxBaseURL = v
	}
	if v := os.Getenv("TORBOX_CACHED_ONLY"); v != "" {
		cfg.TorboxCachedOnly = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}
	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}
}
