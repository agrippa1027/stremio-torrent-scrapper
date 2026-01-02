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

type DatasetEntry struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	InfoHash   string   `json:"infoHash"`
	FileIdx    *int     `json:"fileIdx,omitempty"`
	Size       int64    `json:"size,omitempty"`
	Seeds      int      `json:"seeds,omitempty"`
	Resolution string   `json:"resolution,omitempty"`
	Episode    *Episode `json:"episode,omitempty"`
}

type Episode struct {
	Season  int `json:"season"`
	Episode int `json:"episode"`
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
