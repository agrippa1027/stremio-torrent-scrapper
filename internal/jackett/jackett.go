package jackett

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"scrapper/internal/indexer"

	"github.com/charmbracelet/log"
	jackett "github.com/webtor-io/go-jackett"
)

type Jackett struct {
	client *jackett.Client
	logger *log.Logger
}

func New(client *http.Client, logger *log.Logger) (*Jackett, error) {
	j, err := jackett.New(jackett.Settings{
		ApiURL: os.Getenv("JACKETT_API_URL"),
		ApiKey: os.Getenv("JACKETT_API_KEY"),
		Client: client,
	})

	if err != nil {
		return nil, err
	}

	return &Jackett{client: j, logger: logger}, nil
}

func (j *Jackett) Name() string {
	return "Jackett"
}

func (j *Jackett) Search(ctx context.Context,
	query indexer.SearchQuery) ([]indexer.Torrent, error) {
	var results []jackett.Result
	var err error
	var req *jackett.FetchRequest

	switch query.Type {
	case indexer.QTypeMovie:
		req = jackett.
			NewMovieSearch().
			WithQuery(query.TMDBTitle).
			Build()
	case indexer.QTypeSeries:
		req = jackett.
			NewTVSearch().
			WithQuery(query.TMDBTitle).
			WithSeason(uint(query.Season)).
			WithEpisode(uint(query.Episode)).
			Build()
	default:
		return nil, fmt.Errorf("unsupported search type: %s", query.Type)
	}

	// NOTE:
	// Fetch will concurrently request data from multiple trackers when possible.
	// we can set number of concurrent requests with WithMaxConcurrency
	results, err = j.client.Fetch(ctx, req)

	if err != nil {
		return nil, err
	}

	out := make([]indexer.Torrent, 0, len(results))
	for _, result := range results {
		// WARNING:
		// torrents like 1337x.to does not provide magnet uri in jackett
		// it provides a torrent download link instead
		magnet := result.MagnetURI
		if magnet == "" {
			magnet = result.Link
		}

		out = append(out, indexer.Torrent{
			Title:    result.Title,
			InfoHash: result.InfoHash,
			Magnet:   magnet,
			Size:     result.Size,
			Tracker:  result.Tracker,
			Files:    result.Files,
		})
	}

	return out, nil
}
