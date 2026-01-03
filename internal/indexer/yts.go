package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type YTS struct {
	client *http.Client
}

func NewYTS(client *http.Client) *YTS {
	return &YTS{client: client}
}

func (y *YTS) Name() string {
	return "YTS"
}

func (y *YTS) Search(ctx context.Context, query string) ([]Result, error) {
	endpoint := "https://yts.mx/api/v2/list_movies.json"
	q := url.Values{}
	q.Set("query_term", query)
	q.Set("limit", "10")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	WithUserAgent(req)

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yts status: %s", resp.Status)
	}

	var payload ytsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	results := make([]Result, 0, 10)
	for _, movie := range payload.Data.Movies {
		for _, torrent := range movie.Torrents {
			results = append(results, Result{
				Title:    fmt.Sprintf("%s (%d) %s", movie.Title, movie.Year, strings.ToUpper(torrent.Quality)),
				InfoHash: torrent.Hash,
				Seeds:    torrent.Seeds,
				Size:     torrent.SizeBytes,
				Source:   y.Name(),
				PageURL:  movie.URL,
			})
		}
	}

	return results, nil
}

type ytsResponse struct {
	Data struct {
		Movies []struct {
			Title    string `json:"title"`
			Year     int    `json:"year"`
			URL      string `json:"url"`
			Torrents []struct {
				Hash      string `json:"hash"`
				Quality   string `json:"quality"`
				Seeds     int    `json:"seeds"`
				SizeBytes int64  `json:"size_bytes"`
			} `json:"torrents"`
		} `json:"movies"`
	} `json:"data"`
}
