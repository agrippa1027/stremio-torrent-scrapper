package torbox

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

func (c *Client) TorboxStreamURL(
	ctx context.Context,
	hash string,
	title string,
	cachedOnly bool,
	season int,
	episode int,
	magnet string,
) (string, error) {
	created, err := c.CreateTorrent(ctx, magnet, cachedOnly)
	if err != nil {
		var apiErr APIError
		if errors.As(err, &apiErr) && strings.EqualFold(apiErr.Code, "DUPLICATE_ITEM") {
			existing, findErr := c.FindTorrentByHash(ctx, hash)
			if findErr != nil {
				return "", err
			}
			return c.requestTorboxLink(ctx, existing.ID, title, season, episode)
		}
		return "", err
	}

	return c.requestTorboxLink(ctx, created.TorrentID, title, season, episode)
}

func (c *Client) requestTorboxLink(
	ctx context.Context,
	torrentID int,
	title string,
	season int,
	episode int,
) (string, error) {
	torrent, err := c.GetTorrent(ctx, torrentID)
	if err != nil {
		return "", err
	}
	if !torrent.DownloadPresent {
		return "", errors.New("torrent not ready; try again later")
	}

	file, ok := selectEpisodeFile(torrent.Files, title, season, episode)
	if !ok {
		return "", errors.New("no playable file")
	}

	return c.RequestDownload(ctx, torrentID, file.ID)
}

func selectEpisodeFile(files []TorrentFile, title string, season int, episode int) (TorrentFile, bool) {
	if season > 0 {
		if match, ok := matchEpisodeFile(files, title, season, episode); ok {
			return match, true
		}
	}

	var best TorrentFile
	for _, file := range files {
		if !isVideoFile(file.ShortName) {
			continue
		}
		if file.Size > best.Size {
			best = file
		}
	}
	if best.ID == 0 {
		return TorrentFile{}, false
	}
	return best, true
}

var episodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)s(\d{1,2})e(\d{1,2})`),
	regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,2})`),
	regexp.MustCompile(`(?i)season[ ._-]?(\d{1,2}).*episode[ ._-]?(\d{1,2})`),
}

func matchEpisodeFile(files []TorrentFile, title string, season int, episode int) (TorrentFile, bool) {
	query := strings.ToLower(title)
	var best TorrentFile

	for _, file := range files {
		name := strings.ToLower(file.Name + " " + file.ShortName)
		if !isVideoFile(file.ShortName) {
			continue
		}
		fileSeason, fileEpisode, ok := parseEpisodeMarker(name)
		if !ok || fileSeason != season || fileEpisode != episode {
			continue
		}
		if query != "" && !strings.Contains(name, query) {
			continue
		}
		if file.Size > best.Size {
			best = file
		}
	}

	if best.ID == 0 {
		return TorrentFile{}, false
	}

	return best, true
}

func parseEpisodeMarker(name string) (int, int, bool) {
	for _, re := range episodePatterns {
		matches := re.FindStringSubmatch(name)
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

func isVideoFile(name string) bool {
	name = strings.ToLower(name)
	return strings.HasSuffix(name, ".mkv") ||
		strings.HasSuffix(name, ".mp4") ||
		strings.HasSuffix(name, ".avi") ||
		strings.HasSuffix(name, ".mov")
}

func parseBool(raw string, fallback bool) bool {
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err == nil {
		return value
	}
	return fallback
}

func parseInt(raw string) int {
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return value
}
