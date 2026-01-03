package indexer

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	jackett "github.com/webtor-io/go-jackett"
)

type Jackett struct {
	client *jackett.Client
}

func NewJackett(client *http.Client, apiURL string, apiKey string) (*Jackett, error) {
	j, err := jackett.New(jackett.Settings{
		ApiURL: apiURL,
		ApiKey: apiKey,
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

func (j *Jackett) Search(ctx context.Context, query SearchQuery) ([]Result, error) {
	imdbID := normalizeIMDBID(query.ID)
	var req *jackett.FetchRequest
	if imdbID != "" {
		switch query.Type {
		case "movie":
			req = jackett.NewMovieSearch().
				WithIMDBID(imdbID).
				Build()
		case "series":
			tv := jackett.NewTVSearch().
				WithIMDBID(imdbID)
			if query.Season > 0 {
				tv = tv.WithSeason(uint(query.Season))
			}
			if query.Episode > 0 {
				tv = tv.WithEpisode(uint(query.Episode))
			}
			req = tv.Build()
		}
	}
	if req == nil {
		req = jackett.NewRawSearch().
			WithQuery(query.Query).
			Build()
	}

	results, err := j.client.Fetch(ctx, req)
	if err != nil {
		return nil, err
	}

	out := make([]Result, 0, len(results))
	for _, result := range results {
		out = append(out, Result{
			Title:    result.Title,
			InfoHash: result.InfoHash,
			Magnet:   result.MagnetURI,
			Size:     int64(result.Size),
			Seeds:    int(result.Seeders),
			Source:   result.Tracker,
			PageURL:  result.Link,
		})
	}

	return out, nil
}

var imdbPattern = regexp.MustCompile(`^tt\\d{7,}$`)

func normalizeIMDBID(id string) string {
	id = strings.TrimSpace(id)
	if imdbPattern.MatchString(id) {
		return id
	}
	return ""
}
