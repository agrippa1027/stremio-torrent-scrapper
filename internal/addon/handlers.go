package addon

import "net/http"

func (a *Addon) manifestHandler(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("manifest request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote", r.RemoteAddr,
	)
	writeJSON(w, a.manifest)
}

func (a *Addon) streamHandler(w http.ResponseWriter, r *http.Request) {
	typeStr, idStr, ok := parseStreamPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	baseID, season, episode := splitID(idStr)
	query := buildQuery(typeStr, baseID, season, episode)
	a.logger.Info("stream request",
		"type", typeStr,
		"id", baseID,
		"season", season,
		"episode", episode,
		"query", query,
		"remote", r.RemoteAddr,
	)
	results := a.searchStreams(r.Context(), query)
	streams := buildStreams(results, a.trackers)
	writeJSON(w, StreamResponse{Streams: streams})
}
