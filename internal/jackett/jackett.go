package jackett

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"scrapper/infrastructure"

	"scrapper/internal/indexer"

	jackett "github.com/webtor-io/go-jackett"
)

type Jackett struct {
	client *jackett.Client
}

func NewJackett(client *http.Client) (*Jackett, error) {
	j, err := jackett.New(jackett.Settings{
		ApiURL: os.Getenv("JACKETT_API_URL"),
		ApiKey: os.Getenv("JACKETT_API_KEY"),
		Client: client,
	})

	if err != nil {
		return nil, err
	}

	return &Jackett{client: j}, nil
}

func (j *Jackett) Name() string {
	return "Jackett"
}

func (j *Jackett) Search(ctx context.Context, query indexer.SearchQuery) ([]indexer.Torrent, error) {
	logger := infrastructure.FromContext(ctx)
	logger.Debug("Jackett search",
		"imdb_id", query.ID,
		"type", query.Type,
		"season", query.Season,
		"episode", query.Episode,
	)

	var results []jackett.Result
	var err error

	// TODO: should do the search by tracker.
	// this would allow us to utilize concurrent search
	switch query.Type {
	case "movie":
		req := jackett.NewMovieSearch().WithIMDBID(query.ID).Build()
		results, err = j.client.Fetch(ctx, req)
	case "series":
		results, err = j.searchSeries(ctx, query, query.ID)
	default:
		return nil, fmt.Errorf("unsupported search type: %s", query.Type)
	}

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
