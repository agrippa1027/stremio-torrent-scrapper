package addon

import "net/http"

func (a *Addon) manifestHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.manifest)
}

func (a *Addon) streamHandler(w http.ResponseWriter, r *http.Request) {
	typeStr, idStr, ok := parseStreamPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	baseID, season, episode := splitID(idStr)
	streams := buildStreams(a.dataset, a.trackers, typeStr, baseID, season, episode)
	writeJSON(w, StreamResponse{Streams: streams})
}
