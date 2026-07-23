package stremio

import (
	"fmt"
	"net/http"
	"strings"

	"scrapper/internal/indexer"
	"scrapper/internal/stream"
	"scrapper/internal/util"
)

func (a *Addon) manifestHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: @cors
	// why is this here. do we realy need this?
	util.SetCORSHeaders(w)

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
	// TODO: @cors
	util.SetCORSHeaders(w)

	a.logger.Debug("stream handler invoked",
		"method", r.Method,
		"url", r.URL.Path,
		"remote", r.RemoteAddr,
	)

	searchQuery, err := searchQueryFromRequest(r)
	if err != nil {
		a.logger.Error("search query error", "error", err)
		http.NotFound(w, r)
		return
	}

	results := stream.SearchStreams(r.Context(), searchQuery)

	a.logger.Debug("stream results", "len:", len(results))

	streams := a.buildTorboxStreamLinks(newRequestContext(r), results, typeStr, season, episode)

	writeJSON(w, StreamResponse{Streams: streams})
}

func searchQueryFromRequest(r *http.Request) (indexer.SearchQuery, error) {
	typeStr := r.PathValue("type")
	idRaw := r.PathValue("id")
	idStr, ok := strings.CutSuffix(idRaw, ".json")

	if typeStr == "" || !ok || idStr == "" {
		return indexer.SearchQuery{}, fmt.Errorf("invalid type: %s, id: %s", typeStr, idStr)
	}

	baseID, season, episode := util.SplitID(idStr)

	return indexer.SearchQuery{
		Type:    typeStr,
		ID:      baseID,
		Season:  season,
		Episode: episode,
	}, nil
}
