package stream

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"scrapper/internal/jackett"
)

func (a *Addon) buildTorboxStreamLinks(ctx *requestContext, results []jackett.Torrent, typ string, season int, episode int) []Stream {
	a.logger.Info("building torbox stream links",
		"numResults", len(results),
	)

	const maxResults = 30
	filtered := filterEpisodeResults(results, typ, season, episode)
	sorted := sortResultsByQuality(filtered)
	streams := make([]Stream, 0, maxResults)

	for _, result := range sorted {
		if len(streams) >= maxResults {
			break
		}
		infoHash := result.InfoHash
		if infoHash == "" {
			infoHash = infoHashFromMagnet(result.Magnet)
		}
		if infoHash == "" {
			continue
		}

		cached, err := a.torbox.CheckCached(ctx.ctx, infoHash)
		if err != nil {
			a.logger.Warn("torbox checkcached failed",
				"hash", infoHash,
				"error", err,
			)
			continue
		}

		a.logger.Debug("torbox checkcached result",
			"isCached", cached,
		)

		label := "Torbox (add)"
		link := fmt.Sprintf("%s/torbox/stream/%s?cached=0&title=%s",
			ctx.baseURL,
			infoHash,
			url.QueryEscape(result.Title),
		)
		if typ == "series" && season > 0 && episode > 0 {
			link += fmt.Sprintf("&season=%d&episode=%d", season, episode)
		}
		if cached {
			label = "Torbox (cached)"
			directURL, err := a.torboxStreamURL(ctx.ctx, infoHash, result.Title, true, season, episode)
			if err != nil {
				a.logger.Warn("torbox cached link failed",
					"hash", infoHash,
					"error", err,
				)
			} else {
				a.logger.Info("torbox checkcached result",
					"cached", cached,
					"directURL", directURL,
				)
				link = directURL
			}
		}

		a.logger.Info("built torbox stream link",
			"title", result.Title,
			"link", link,
			"info_hash", infoHash,
		)

		streams = append(streams, Stream{
			Name:     label,
			Title:    formatTitle(result),
			URL:      link,
			InfoHash: infoHash,
		})
	}

	return streams
}

func buildMagnet(infoHash string, title string) string {
	vals := url.Values{}
	vals.Set("xt", "urn:btih:"+strings.ToUpper(infoHash))
	if title != "" {
		vals.Set("dn", title)
	}
	return "magnet:?" + vals.Encode()
}

var episodeRegexes = []*regexp.Regexp{
	regexp.MustCompile(`(?i)s(\d{1,2})e(\d{1,2})`),
	regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,2})`),
	regexp.MustCompile(`(?i)season[ ._-]?(\d{1,2}).*episode[ ._-]?(\d{1,2})`),
}

func filterEpisodeResults(results []jackett.Torrent, typ string, season int, episode int) []jackett.Torrent {
	if typ != "series" || season <= 0 || episode <= 0 {
		return results
	}

	filtered := make([]jackett.Torrent, 0, len(results))
	for _, result := range results {
		matchSeason, matchEpisode, ok := extractEpisodeMarker(result.Title)
		if !ok || matchSeason != season || matchEpisode != episode {
			continue
		}
		filtered = append(filtered, result)
	}

	if len(filtered) == 0 {
		return results
	}
	return filtered
}

func extractEpisodeMarker(title string) (int, int, bool) {
	candidate := strings.ToLower(title)
	for _, re := range episodeRegexes {
		matches := re.FindStringSubmatch(candidate)
		if len(matches) != 3 {
			continue
		}
		season, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		episode, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}
		return season, episode, true
	}
	return 0, 0, false
}

func sortResultsByQuality(results []jackett.Torrent) []jackett.Torrent {
	sorted := make([]jackett.Torrent, len(results))
	copy(sorted, results)
	if len(sorted) < 2 {
		return sorted
	}

	sort.SliceStable(sorted, func(i, j int) bool {
		return qualityRank(sorted[i].Title) > qualityRank(sorted[j].Title)
	})
	return sorted
}

func qualityRank(title string) int {
	candidate := strings.ToLower(title)
	score := 0
	if strings.Contains(candidate, "2160p") || strings.Contains(candidate, "4k") || strings.Contains(candidate, "uhd") {
		score += 400
	} else if strings.Contains(candidate, "1080p") {
		score += 300
	} else if strings.Contains(candidate, "720p") {
		score += 200
	} else if strings.Contains(candidate, "480p") {
		score += 100
	}

	if strings.Contains(candidate, "bluray") || strings.Contains(candidate, "blu-ray") || strings.Contains(candidate, "bdrip") || strings.Contains(candidate, "brrip") {
		score += 80
	} else if strings.Contains(candidate, "web-dl") || strings.Contains(candidate, "webdl") || strings.Contains(candidate, "webrip") {
		score += 60
	} else if strings.Contains(candidate, "hdtv") {
		score += 20
	}

	if strings.Contains(candidate, "x265") || strings.Contains(candidate, "hevc") {
		score += 20
	} else if strings.Contains(candidate, "x264") {
		score += 10
	}

	return score
}

func infoHashFromMagnet(magnet string) string {
	if magnet == "" {
		return ""
	}
	const prefix = "magnet:?"
	if strings.HasPrefix(magnet, prefix) {
		magnet = magnet[len(prefix):]
	}
	values, err := url.ParseQuery(magnet)
	if err != nil {
		return ""
	}
	xt := strings.ToLower(values.Get("xt"))
	const urnPrefix = "urn:btih:"
	if !strings.HasPrefix(xt, urnPrefix) {
		return ""
	}
	return strings.ToUpper(strings.TrimPrefix(xt, urnPrefix))
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

func formatTitle(result jackett.Torrent) string {
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

func humanSize(size uint64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := uint64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
