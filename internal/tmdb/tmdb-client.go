package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func New(opts Options, httpClient *http.Client) (*Client, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("missing TMDB API key")
	}

	parsed, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid TMDB base URL: %w", err)
	}

	if httpClient == nil {
		return nil, fmt.Errorf("http client is required")
	}

	return &Client{
		host:    parsed,
		key:     opts.APIKey,
		http:    httpClient,
		timeout: time.Duration(opts.TimeoutSeconds) * time.Second,
	}, nil
}

func (c *Client) TitleByIMDB(ctx context.Context, imdbID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	imdbID = strings.TrimSpace(imdbID)
	if imdbID == "" {
		return "", fmt.Errorf("imdb id is required")
	}

	findRequest, err := c.newFindRequest(ctx, imdbID)
	if err != nil {
		return "", fmt.Errorf("failed to create find request: %w", err)
	}

	resp, err := c.http.Do(findRequest)
	if err != nil {
		return "", fmt.Errorf("failed to send find request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tmdb status: %s", resp.Status)
	}

	var result findResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode TMDB find response: %w", err)
	}

	return result.title(), nil

}

func (c *Client) newFindRequest(
	ctx context.Context,
	imdbID string,
) (*http.Request, error) {
	endpoint := *c.host
	endpoint.Path = path.Join(endpoint.Path, "find", imdbID)

	query := endpoint.Query()
	query.Set("external_source", "imdb_id")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create TMDB find request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (r findResponse) title() string {
	if len(r.MovieResults) > 0 {
		return r.MovieResults[0].Title
	}
	if len(r.TVResults) > 0 {
		return r.TVResults[0].Title
	}
	return ""
}
