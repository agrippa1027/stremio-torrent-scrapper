package addon

type Config struct {
	Manifest         Manifest `yaml:"-"`
	JackettURL       string   `yaml:"jacketturl"`
	JackettAPIKey    string   `yaml:"jackettapikey"`
	TMDBAPIKey       string   `yaml:"tmdbapikey"`
	Port             string   `yaml:"port"`
	TorboxAPIKey     string   `yaml:"torboxapikey"`
	TorboxBaseURL    string   `yaml:"torboxbaseurl"`
	TorboxCachedOnly bool     `yaml:"torboxcachedonly"`
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
		Port:             "8000",
		TorboxBaseURL:    "https://api.torbox.app/v1/api",
		TorboxCachedOnly: true,
	}
}
