package addon

type Manifest struct {
	ID          string   `json:"id"`
	Version     string   `json:"version"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resources   []string `json:"resources"`
	Types       []string `json:"types"`
	Catalogs    []any    `json:"catalogs"`
}

type Stream struct {
	Name     string   `json:"name,omitempty"`
	Title    string   `json:"title,omitempty"`
	URL      string   `json:"url,omitempty"`
	InfoHash string   `json:"infoHash,omitempty"`
	FileIdx  *int     `json:"fileIdx,omitempty"`
	Sources  []string `json:"sources,omitempty"`
}

type StreamResponse struct {
	Streams []Stream `json:"streams"`
}
