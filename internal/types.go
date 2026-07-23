package addon

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
