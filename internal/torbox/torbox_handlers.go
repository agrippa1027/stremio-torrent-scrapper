package torbox

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"scrapper/internal/torbox"
	"strconv"
	"strings"
)

func (a *Addon) torboxStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	hashRaw := strings.TrimSpace(r.PathValue("hash"))
	hash, ok := strings.CutSuffix(hashRaw, ".json")
	if !ok || hash == "" {
		http.NotFound(w, r)
		return
	}

	cachedOnly := parseBool(r.URL.Query().Get("cached"), a.torboxCachedOnly)
	title := r.URL.Query().Get("title")
	season := parseInt(r.URL.Query().Get("season"))
	episode := parseInt(r.URL.Query().Get("episode"))

	ctx := r.Context()
	url, err := a.torboxStreamURL(ctx, hash, title, cachedOnly, season, episode)
	if err != nil {
		a.logger.Warn("torbox stream failed",
			"hash", hash,
			"cached_only", cachedOnly,
			"error", err,
		)
		status := http.StatusBadRequest
		var apiErr *torbox.APIError
		if errors.As(err, &apiErr) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (a *Addon) torboxStreamURL(ctx context.Context, hash string, title string, cachedOnly bool, season int, episode int) (string, error) {
	magnet := buildMagnet(hash, title)

	created, err := a.torbox.CreateTorrent(ctx, magnet, cachedOnly)
	if err != nil {
		var apiErr *torbox.APIError
		if errors.As(err, &apiErr) && strings.EqualFold(apiErr.Code, "DUPLICATE_ITEM") {
			existing, findErr := a.torbox.FindTorrentByHash(ctx, hash)
			if findErr != nil {
				return "", err
			}
			return a.requestTorboxLink(ctx, existing.ID, title, season, episode)
		}
		return "", err
	}
	a.logger.Debug("torbox create torrent",
		"createdID", created.TorrentID,
	)

	return a.requestTorboxLink(ctx, created.TorrentID, title, season, episode)
}

func (a *Addon) requestTorboxLink(ctx context.Context, torrentID int, title string, season int, episode int) (string, error) {
	torrent, err := a.torbox.GetTorrent(ctx, torrentID)
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

	return a.torbox.RequestDownload(ctx, torrentID, file.ID)
}

func selectEpisodeFile(files []torbox.TorrentFile, title string, season int, episode int) (torbox.TorrentFile, bool) {
	if season > 0 {
		if match, ok := matchEpisodeFile(files, title, season, episode); ok {
			return match, true
		}
	}

	var best torbox.TorrentFile
	for _, file := range files {
		if !isVideoFile(file.ShortName) {
			continue
		}
		if file.Size > best.Size {
			best = file
		}
	}
	if best.ID == 0 {
		return torbox.TorrentFile{}, false
	}
	return best, true
}

var episodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)s(\d{1,2})e(\d{1,2})`),
	regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,2})`),
	regexp.MustCompile(`(?i)season[ ._-]?(\d{1,2}).*episode[ ._-]?(\d{1,2})`),
}

func matchEpisodeFile(files []torbox.TorrentFile, title string, season int, episode int) (torbox.TorrentFile, bool) {
	query := strings.ToLower(title)
	var best torbox.TorrentFile

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
		return torbox.TorrentFile{}, false
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
