package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"scrapper/internal/config"
)

var ErrMissingAPIKey = errors.New("missing TMDB API key")


type Client struct {
	baseURL *url.URL
	apiKey  string
	http    *http.Client
}

type TVResults struct {
	Name string `json:"name"`
}


func NewClient(tmdbCfg config.TMDBConfig, httpClient *http.Client) (*Client, error) {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	parsed, err := url.Parse(tmdbCfg.BaseURL)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Duration(tmdbCfg.Timeout) * time.Second,
		}
	}

	return &Client{
		baseURL: parsed,
		apiKey:  apiKey,
		http:    httpClient,
	}, nil
}

func (c *Client) TVTitleByIMDB(ctx context.Context, imdbID string) ([]TVResults, error) {
	imdbID = strings.TrimSpace(imdbID)
	if imdbID == "" {
		return nil, errors.New("missing imdb id")
	}

	req, err := c.buildRequest(ctx, http.MethodGet, imdbID)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tmdb status: %s", resp.Status)
	}

	var payload struct {
		TVResults []TVResults `json:"tv_results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}


	return payload.TVResults, nil
}


func (c *Client) buildRequest(ctx context.Context, method, imdbID string) (*http.Request, error) {
	u := *c.baseURL

	u.Path = path.Join(u.Path, "find", imdbID)

	q := u.Query()
	q.Set("external_source", "imdb_id")

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer " + c.apiKey)
	return req, nil
}
