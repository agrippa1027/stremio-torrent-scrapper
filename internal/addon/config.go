package addon

type Config struct {
	Manifest      Manifest `yaml:"-"`
	JackettURL    string   `yaml:"jacketturl"`
	JackettAPIKey string   `yaml:"jackettapikey"`
	Port          string   `yaml:"port"`
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
			Logo:        "https://www.stremio.com/website/stremio-logo-small.png",
		},
		Port: "8000",
	}
}
