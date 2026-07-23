package util

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

/*
* For `series` type, the ID format is expected to be `imdbID:season:episode`, e.g. `tt1234567:2:5`.
* movie type should just be the imdbID, e.g. `tt1234567`.
 */
func SplitID(id string) (string, int, int) {
	parts := strings.Split(id, ":")
	if len(parts) < 3 {
		return id, 0, 0
	}
	season, _ := strconv.Atoi(parts[1])
	episode, _ := strconv.Atoi(parts[2])
	return parts[0], season, episode
}

func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

type requestContext struct {
	ctx     context.Context
	baseURL string
}

func newRequestContext(r *http.Request) *requestContext {
	return &requestContext{
		ctx:     r.Context(),
		baseURL: requestBaseURL(r),
	}
}

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	return scheme + "://" + r.Host
}
