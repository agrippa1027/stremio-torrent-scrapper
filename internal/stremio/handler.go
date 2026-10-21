package stremio

import (
	"net/http"
	"strings"

	"scrapper/internal/stream"
	"scrapper/internal/util"
)

func (a *Addon) manifestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	a.logger.Info("manifest request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote", r.RemoteAddr,
	)

	util.WriteJSON(w, a.manifest)
}

// Sample request:
// GET stream/series/tt0386676:05:16.json
func (a *Addon) streamHandler(w http.ResponseWriter, r *http.Request) {
	typeStr := r.PathValue("type")
	idRaw := r.PathValue("id")

	idStr, ok := strings.CutSuffix(idRaw, ".json")
	if !ok {
		http.NotFound(w, r)
		a.logger.Warn("invalid id", "id", idRaw)
		return
	}
	imdbID, season, episode := util.SplitID(idStr)
	results, err := a.searchService.Search(
		r.Context(),
		typeStr,
		imdbID,
		season,
		episode,
	)
	if err != nil {
		a.logger.Error("search error", "error", err)
		http.NotFound(w, r)
		return
	}

	if len(results) == 0 {
		a.logger.Warn("no results", "type", typeStr, "id", idStr)
		http.NotFound(w, r)
		return
	}

	util.WriteJSON(w, stream.StreamResponse{Streams: results})
}
