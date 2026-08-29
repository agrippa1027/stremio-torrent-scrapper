package stremio

import (
	"fmt"
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
	searchQuery, err := a.query(r)
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

func (a *Addon) query(r *http.Request) (stream.SearchQuery, error) {
	query, err := parseRequest(r)
	if err != nil {
		return stream.SearchQuery{}, err
	}

	title, err := a.tmdbClient.Title(r.Context(), query.ID)
	if err != nil {
		return stream.SearchQuery{}, err
	}
	query.Title = title

	return query, nil
}

func parseRequest(r *http.Request) (stream.SearchQuery, error) {
	typeStr := r.PathValue("type")
	idRaw := r.PathValue("id")
	idStr, ok := strings.CutSuffix(idRaw, ".json")

	if typeStr == "" || !ok || idStr == "" {
		return stream.SearchQuery{}, fmt.Errorf("invalid type: %s, id: %s", typeStr, idStr)
	}

	baseID, season, episode := util.SplitID(idStr)

	return stream.SearchQuery{
		Type:    typeStr,
		ID:      baseID,
		Season:  season,
		Episode: episode,
	}, nil

}
