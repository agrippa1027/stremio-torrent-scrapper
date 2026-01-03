package addon

import (
	"fmt"
	"net/url"
	"strings"

	"scrapper/internal/indexer"
)

func buildStreams(results []indexer.Result, trackers []string) []Stream {
	streams := make([]Stream, 0, len(results))
	for _, result := range results {
		magnet := result.Magnet
		if magnet == "" && result.InfoHash != "" {
			magnet = buildMagnet(result.InfoHash, result.Title, trackers)
		}

		if magnet == "" {
			continue
		}

		streams = append(streams, Stream{
			Name:     result.Source,
			Title:    formatTitle(result),
			URL:      magnet,
			InfoHash: result.InfoHash,
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

func formatTitle(result indexer.Result) string {
	var parts []string
	if result.Size > 0 {
		parts = append(parts, humanSize(result.Size))
	}
	if result.Seeds > 0 {
		parts = append(parts, fmt.Sprintf("%d seeds", result.Seeds))
	}
	if result.Source != "" {
		parts = append(parts, result.Source)
	}
	if len(parts) == 0 {
		return result.Title
	}
	return fmt.Sprintf("%s • %s", result.Title, strings.Join(parts, " • "))
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
