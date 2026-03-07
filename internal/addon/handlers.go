package addon

import (
	"net/http"

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

func (a *Addon) streamHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	a.logger.Debug("stream handler invoked", "url", r.URL.Path)
	typeStr, idStr, ok := parseStreamPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	baseID, season, episode := splitID(idStr)
	query := buildQuery(typeStr, baseID, season, episode)

	a.logger.Debug("stream request",
		"type", typeStr,
		"id", baseID,
		"season", season,
		"episode", episode,
		"query", query,
		"remote", r.RemoteAddr,
	)
	results := a.searchStreams(r.Context(), indexer.SearchQuery{
		Type:    typeStr,
		ID:      baseID,
		Query:   query,
		Season:  season,
		Episode: episode,
	})
	a.logger.Info("stream results", "len:", len(results))
	streams := a.buildTorboxStreamLinks(newRequestContext(r), results, typeStr, season, episode)
	writeJSON(w, StreamResponse{Streams: streams})
}
