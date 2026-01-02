package addon

import "path/filepath"

type Config struct {
	DatasetPath  string
	TrackersPath string
	Manifest     Manifest
}

func DefaultConfig() Config {
	return Config{
		DatasetPath:  filepath.Join("data", "dataset.json"),
		TrackersPath: filepath.Join("data", "trackers.txt"),
		Manifest: Manifest{
			ID:          "com.local.dataset.torrents",
			Version:     "0.1.0",
			Name:        "Local Dataset Torrents",
			Description: "Streams only from your local dataset using public torrent trackers.",
			Resources:   []string{"stream"},
			Types:       []string{"movie", "series"},
			Catalogs:    []any{},
		},
	}
}
