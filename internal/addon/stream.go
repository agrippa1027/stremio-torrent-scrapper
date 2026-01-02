package addon

import (
	"fmt"
	"net/url"
	"strings"
)

func buildStreams(entries []DatasetEntry, trackers []string, typ string, id string, season int, episode int) []Stream {
	var streams []Stream
	for _, entry := range entries {
		if entry.Type != typ || entry.ID != id {
			continue
		}
		if entry.Episode != nil {
			if season == 0 || episode == 0 {
				continue
			}
			if entry.Episode.Season != season || entry.Episode.Episode != episode {
				continue
			}
		}

		magnet := buildMagnet(entry.InfoHash, entry.Title, trackers)
		streams = append(streams, Stream{
			Name:     "Local Dataset",
			Title:    formatTitle(entry),
			URL:      magnet,
			InfoHash: entry.InfoHash,
			FileIdx:  entry.FileIdx,
			Sources:  trackerSources(trackers),
		})
	}

	return streams
}

func buildMagnet(infoHash string, title string, trackers []string) string {
	vals := url.Values{}
	vals.Set("xt", "urn:btih:"+strings.ToUpper(infoHash))
	if title != "" {
		vals.Set("dn", title)
	}
	for _, tr := range trackers {
		vals.Add("tr", tr)
	}
	return "magnet:?" + vals.Encode()
}

func trackerSources(trackers []string) []string {
	if len(trackers) == 0 {
		return nil
	}

	out := make([]string, 0, len(trackers))
	for _, tr := range trackers {
		out = append(out, "tracker:"+tr)
	}
	return out
}

func formatTitle(entry DatasetEntry) string {
	var parts []string
	if entry.Resolution != "" {
		parts = append(parts, entry.Resolution)
	}
	if entry.Size > 0 {
		parts = append(parts, humanSize(entry.Size))
	}
	if entry.Seeds > 0 {
		parts = append(parts, fmt.Sprintf("%d seeds", entry.Seeds))
	}
	if len(parts) == 0 {
		return entry.Title
	}
	return fmt.Sprintf("%s • %s", entry.Title, strings.Join(parts, " • "))
}

func humanSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
