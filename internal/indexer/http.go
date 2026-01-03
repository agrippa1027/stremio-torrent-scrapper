package indexer

import (
	"net/http"
	"time"
)

func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 12 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

func WithUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; StremioAddon/1.0)")
}
