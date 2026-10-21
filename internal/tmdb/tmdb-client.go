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

	"github.com/redis/go-redis/v9"
)

func New(opts Options, httpClient *http.Client, redisClient *redis.Client, logger *Logger) (*Client, error) {
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
		redis:   redisClient,
		timeout: time.Duration(opts.TimeoutSeconds) * time.Second,
		logger:  logger,
	}, nil
}

func (c *Client) TitleByIMDB(ctx context.Context, imdbID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	imdbID = strings.TrimSpace(imdbID)
	if imdbID == "" {
		return "", fmt.Errorf("[TMDB][TitleByIMDB] imdb id is required")
	}

	cachedResult, err := c.redis.Get(ctx, imdbID).Result()
	if err != nil {
		return "", fmt.Errorf("[TMDB][TitleByIMDB] failed to get cached title: %w", err)
	}
	if cachedResult != "" {
		c.logger.Info("[TMDB][TitleByIMDB] using cached title: ", "imdbID", imdbID, "title", cachedResult)
		return cachedResult, nil
	}

	findRequest, err := c.newFindRequest(ctx, imdbID)
	if err != nil {
		return "", fmt.Errorf("[TMDB][TitleByIMDB] failed to create find request: %w", err)
	}

	resp, err := c.http.Do(findRequest)
	if err != nil {
		return "", fmt.Errorf("[TMDB][TitleByIMDB] failed to send find request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("[TMDB][TitleByIMDB] status: %s for request: %s", resp.Status, findRequest.URL)
	}

	var result findResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("[TMDB][TitleByIMDB] decode TMDB find response: %w", err)
	}

	title := result.title()

	if err := c.redis.Set(ctx, imdbID, title, 0).Err(); err != nil {
		c.logger.Error("[TMDB][TitleByIMDB] failed to set cached title: ", "error", err)
	}

	return title, nil
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
