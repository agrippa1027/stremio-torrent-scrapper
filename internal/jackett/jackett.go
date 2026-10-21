package jackett

import (
	"context"
	"fmt"
	"net/http"

	jackett "github.com/webtor-io/go-jackett"
)

type Client struct {
	client *jackett.Client
}

func New(opt Options, httpClient *http.Client) (*Client, error) {
	j, err := jackett.New(jackett.Settings{
		ApiURL: opt.BaseURL,
		ApiKey: opt.APIKey,
		Client: httpClient,
	})

	if err != nil {
		return nil, err
	}

	return &Client{client: j}, nil
}

func (j *Client) Search(ctx context.Context, query SearchQuery) ([]Torrent, error) {
	var results []jackett.Result
	var err error
	var req *jackett.FetchRequest

	switch query.Type {
	case QTypeMovie:
		req = jackett.
			NewMovieSearch().
			WithQuery(query.TMDBTitle).
			Build()
	case QTypeSeries:
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

	out := make([]Torrent, 0, len(results))
	for _, result := range results {
		// WARNING:
		// torrents like 1337x.to does not provide magnet uri in jackett
		// it provides a torrent download link instead
		magnet := result.MagnetURI
		if magnet == "" {
			magnet = result.Link
		}

		out = append(out, Torrent{
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
