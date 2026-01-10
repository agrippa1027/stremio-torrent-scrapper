package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var ErrMissingAPIKey = errors.New("missing TMDB API key")

type Client struct {
	baseURL *url.URL
	apiKey  string
	http    *http.Client
}

func NewClient(apiKey string, httpClient *http.Client) (*Client, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	parsed, err := url.Parse("https://api.themoviedb.org/3")
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{baseURL: parsed, apiKey: apiKey, http: httpClient}, nil
}

func (c *Client) TVTitleByIMDB(ctx context.Context, imdbID string) (string, error) {
	imdbID = strings.TrimSpace(imdbID)
	if imdbID == "" {
		return "", errors.New("missing imdb id")
	}

	u := *c.baseURL
	u.Path = path.Join(u.Path, "find", imdbID)
	q := u.Query()
	q.Set("external_source", "imdb_id")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tmdb status: %s", resp.Status)
	}

	var payload struct {
		TVResults []struct {
			Name string `json:"name"`
		} `json:"tv_results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.TVResults) == 0 {
		return "", errors.New("tmdb tv result not found")
	}
	return payload.TVResults[0].Name, nil
}
