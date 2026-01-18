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

func (a *Addon) buildTorboxStreamLinks(ctx *requestContext, results []indexer.Result, typ string, season int, episode int) []Stream {
	a.logger.Info("building torbox stream links",
		"numResults", len(results),
	)

	const maxResults = 10
	streams := make([]Stream, 0, maxResults)

	for _, result := range results {
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
