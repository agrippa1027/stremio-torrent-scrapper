package tmdb

import (
	"net/http"
	"net/url"
	"time"

	"charm.land/log/v2"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	host    *url.URL
	key     string
	http    *http.Client
	redis   *redis.Client
	timeout time.Duration
	logger  *log.Logger
}

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
