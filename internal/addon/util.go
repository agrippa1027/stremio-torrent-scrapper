package addon

import (
	"encoding/json"
	"net/http"
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
	return parts[0], parts[1], true
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
