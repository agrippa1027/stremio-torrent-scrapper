package addon

import (
	"net/http"
	"strings"

	"scrapper/internal/indexer"
)

func (a *Addon) manifestHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	a.logger.Info("manifest request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote", r.RemoteAddr,
	)
	writeJSON(w, a.manifest)
}

// GET stream/series/tt0386676:05:16.json
func (a *Addon) streamHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	a.logger.Debug("stream handler invoked", "url", r.URL.Path)

	typeStr := r.PathValue("type")
	idRaw := r.PathValue("id")
	idStr, ok := strings.CutSuffix(idRaw, ".json")

	if typeStr == "" || !ok || idStr == "" {
		http.NotFound(w, r)
		return
	}

	switch typeStr {
	case "movie", "series":
		// valid
	default:
		http.Error(w, "invalid type: "+typeStr, http.StatusBadRequest)
		return
	}

	baseID, season, episode := splitID(idStr)
	query := buildQuery(typeStr, baseID, season, episode)

	a.logger.Debug("stream request",
		"type", typeStr, // series
		"id", baseID, // tt0386676
		"season", season, // 5
		"episode", episode, // 16
		"query", query, // tt0386676 S05E16
		"remote", r.RemoteAddr, // [::1]:62481

	)

	results := a.searchStreams(r.Context(), indexer.SearchQuery{
		Type:    typeStr,
		ID:      baseID,
		Query:   query,
		Season:  season,
		Episode: episode,
	})

	a.logger.Debug("stream results", "len:", len(results))

	streams := a.buildTorboxStreamLinks(newRequestContext(r), results, typeStr, season, episode)

	writeJSON(w, StreamResponse{Streams: streams})
}
