package addon

type Config struct {
	Manifest Manifest
}

func DefaultConfig() Config {
	return Config{
		Manifest: Manifest{
			ID:          "com.public.indexers",
			Version:     "0.1.0",
			Name:        "Public Indexer Streams",
			Description: "Streams from public torrent indexers using title search.",
			Resources:   []string{"stream"},
			Types:       []string{"movie", "series"},
			Catalogs:    []any{},
		},
	}
}
