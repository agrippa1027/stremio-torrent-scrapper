package indexer

import (
	"context"
	"net/http"

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

func (j *Jackett) Search(ctx context.Context, query string) ([]Result, error) {
	req := jackett.NewRawSearch().
		WithQuery(query).
		Build()

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
