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

	"github.com/charmbracelet/log"

	"scrapper/internal/config"
)

var ErrMissingAPIKey = errors.New("missing TMDB API key")

type Client struct {
	baseURL *url.URL
	apiKey  string
	http    *http.Client
	logger  *log.Logger
}

type TVResults struct {
	Name string `json:"name" redis:"name"`
}

func New(tmdbCfg config.TMDBConfig, httpClient *http.Client, logger *log.Logger) (*Client, error) {
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
		logger:  logger,
	}, nil
}

func (c *Client) Title(ctx context.Context, id string) ([]TVResults, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("imdb id is required")
	}

	req, err := c.request(ctx, http.MethodGet, id)
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
	c.logger.Debug("TVTitleByIMDB", "payload", payload)

	return payload.TVResults, nil
}

func (c *Client) request(ctx context.Context, method, id string) (*http.Request, error) {
	u := *c.baseURL

	u.Path = path.Join(u.Path, "find", id)

	q := u.Query()
	q.Set("external_source", "imdb_id")

	u.RawQuery = q.Encode()

	// TODO: This shoudld be request with timeout
	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	return req, nil
}
