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
	APIKey     string `json:"apiKey"`
	CachedOnly bool   `json:"cachedOnly"`
	Timeout    int    `json:"timeout"`
}

type TMDBOptions struct {
	BaseURL        string `yaml:"baseurl"`
	APIKey         string `yaml:"-"`
	TimeoutSeconds int    `yaml:"timeout"`
}

type JackettConfig struct {
	BaseURL string `yaml:"baseurl"`
	APIKey  string `yaml:"-"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DB       int    `yaml:"db"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	Manifest ManifestConfig `yaml:"manifest"`
	Torbox   TorboxConfig   `yaml:"torbox"`
	TMDB     TMDBOptions    `yaml:"tmdb"`
	Jackett  JackettConfig  `yaml:"jackett"`
	Redis    RedisConfig    `yaml:"redis"`
}

// I don't see the other configs
// such jackett
