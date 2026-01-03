package addon

import (
	"context"
	"errors"
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

	ctx := r.Context()
	url, err := a.torboxStreamURL(ctx, hash, title, cachedOnly)
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

func (a *Addon) torboxStreamURL(ctx context.Context, hash string, title string, cachedOnly bool) (string, error) {
	magnet := buildMagnet(hash, title, a.trackers)

	created, err := a.torbox.CreateTorrent(ctx, magnet, cachedOnly)
	if err != nil {
		var apiErr *torbox.APIError
		if errors.As(err, &apiErr) && strings.EqualFold(apiErr.Code, "DUPLICATE_ITEM") {
			existing, findErr := a.torbox.FindTorrentByHash(ctx, hash)
			if findErr != nil {
				return "", err
			}
			return a.requestTorboxLink(ctx, existing.ID)
		}
		return "", err
	}
	a.logger.Debug("torbox create torrent",
		"createdID", created.TorrentID,
	)

	return a.requestTorboxLink(ctx, created.TorrentID)
}

func (a *Addon) requestTorboxLink(ctx context.Context, torrentID int) (string, error) {
	torrent, err := a.torbox.GetTorrent(ctx, torrentID)
	if err != nil {
		return "", err
	}
	if !torrent.DownloadPresent {
		return "", errors.New("torrent not ready; try again later")
	}

	file, ok := selectBestFile(torrent.Files)
	if !ok {
		return "", errors.New("no playable file")
	}

	return a.torbox.RequestDownload(ctx, torrentID, file.ID)
}

func selectBestFile(files []torbox.TorrentFile) (torbox.TorrentFile, bool) {
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
