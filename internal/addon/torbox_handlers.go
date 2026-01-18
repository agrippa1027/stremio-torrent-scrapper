package addon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"scrapper/internal/torbox"
)

func (a *Addon) torboxStreamHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	hash := strings.TrimPrefix(r.URL.Path, "/torbox/stream/")
	hash = strings.TrimSuffix(hash, ".json")
	hash = strings.TrimSpace(hash)
	if hash == "" {
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
	magnet := buildMagnet(hash, title, a.trackers)

	created, err := a.torbox.CreateTorrent(ctx, magnet, cachedOnly)
	if err != nil {
		var apiErr *torbox.APIError
		if errors.As(err, &apiErr) && strings.EqualFold(apiErr.Code, "DUPLICATE_ITEM") {
			existing, findErr := a.torbox.FindTorrentByHash(ctx, hash)
			if findErr != nil {
				return "", err
			}
			return a.requestTorboxLink(ctx, existing.ID, season, episode)
		}
		return "", err
	}
	a.logger.Debug("torbox create torrent",
		"createdID", created.TorrentID,
	)

	return a.requestTorboxLink(ctx, created.TorrentID, season, episode)
}

func (a *Addon) requestTorboxLink(ctx context.Context, torrentID int, season int, episode int) (string, error) {
	torrent, err := a.torbox.GetTorrent(ctx, torrentID)
	if err != nil {
		return "", err
	}
	if !torrent.DownloadPresent {
		return "", errors.New("torrent not ready; try again later")
	}

	file, ok := selectEpisodeFile(torrent.Files, season, episode)
	if !ok {
		return "", errors.New("no playable file")
	}

	return a.torbox.RequestDownload(ctx, torrentID, file.ID)
}

func selectEpisodeFile(files []torbox.TorrentFile, season int, episode int) (torbox.TorrentFile, bool) {
	if season > 0 && episode > 0 {
		if match, ok := matchEpisodeFile(files, season, episode); ok {
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

func isVideoFile(name string) bool {
	name = strings.ToLower(name)
	return strings.HasSuffix(name, ".mkv") ||
		strings.HasSuffix(name, ".mp4") ||
		strings.HasSuffix(name, ".avi") ||
		strings.HasSuffix(name, ".mov")
}

func matchEpisodeFile(files []torbox.TorrentFile, season int, episode int) (torbox.TorrentFile, bool) {
	target := fmt.Sprintf("s%02de%02d", season, episode)
	targetAlt := fmt.Sprintf("%dx%02d", season, episode)

	for _, file := range files {
		name := strings.ToLower(file.Name + " " + file.ShortName)
		if strings.Contains(name, target) || strings.Contains(name, targetAlt) {
			if isVideoFile(file.ShortName) {
				return file, true
			}
		}
	}

	return torbox.TorrentFile{}, false
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
