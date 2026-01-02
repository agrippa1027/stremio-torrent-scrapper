package addon

import "net/http"

type Addon struct {
	dataset  []DatasetEntry
	trackers []string
	manifest Manifest
}

func New(cfg Config) (*Addon, error) {
	dataset, err := LoadDataset(cfg.DatasetPath)
	if err != nil {
		return nil, err
	}

	trackers, err := LoadTrackers(cfg.TrackersPath)
	if err != nil || len(trackers) == 0 {
		trackers = DefaultTrackers()
	}

	return &Addon{
		dataset:  dataset,
		trackers: trackers,
		manifest: cfg.Manifest,
	}, nil
}

func (a *Addon) Register(mux *http.ServeMux) {
	mux.HandleFunc("/manifest.json", a.manifestHandler)
	mux.HandleFunc("/stream/", a.streamHandler)
}
