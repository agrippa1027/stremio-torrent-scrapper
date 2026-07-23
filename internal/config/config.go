package config

type ManifestConfig struct {
	ID          string   `json:"id"`
	Version     string   `json:"version"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resources   []string `json:"resources"`
	Types       []string `json:"types"`
	Catalogs    []any    `json:"catalogs"`
	Logo        string   `json:"logo,omitempty"`
}

type TorboxConfig struct {
	BaseURL    string `json:"baseUrl"`
	CachedOnly bool   `json:"cachedOnly"`
	Timeout    int    `json:"timeout"`
}

type TMDBConfig struct {
	BaseURL string `yaml:"baseurl"`
	Timeout int    `yaml:"timeout"`
}

type Config struct {
	Manifest ManifestConfig `yaml:"manifest"`
	Torbox   TorboxConfig   `yaml:"torbox"`
	TMDB     TMDBConfig     `yaml:"tmdb"`
}
