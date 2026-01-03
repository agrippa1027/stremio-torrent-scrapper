package addon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func parseStreamPath(path string) (string, string, bool) {
	trimmed := strings.TrimPrefix(path, "/stream/")
	trimmed = strings.TrimSuffix(trimmed, ".json")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	id, err := url.PathUnescape(parts[1])
	if err != nil {
		return "", "", false
	}
	return parts[0], id, true
}

func splitID(id string) (string, int, int) {
	parts := strings.Split(id, ":")
	if len(parts) < 3 {
		return id, 0, 0
	}
	season, _ := strconv.Atoi(parts[1])
	episode, _ := strconv.Atoi(parts[2])
	return parts[0], season, episode
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func setCORSHeaders(w http.ResponseWriter) {
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
