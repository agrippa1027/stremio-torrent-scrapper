package tmdb

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	host    *url.URL
	key     string
	http    *http.Client
	timeout time.Duration
}

var (
	ErrMissingAPIKey = errors.New("missing TMDB API key")
	ErrHTTPClient    = errors.New("http client is required")
	ErrTitleNotFound = errors.New("title not found")
)

type Options struct {
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
}

type findResponse struct {
	MovieResults []movieResult `json:"movie_results"`
	TVResults    []tvResult    `json:"tv_results"`
}

type movieResult struct {
	Title string `json:"title"`
}

type tvResult struct {
	Title string `json:"name"`
}
