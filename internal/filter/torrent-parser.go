package filter

import (
	"context"
	"scrapper/infrastructure"

	ptt "github.com/MunifTanjim/go-ptt"
)

// This fucntion should produce the actual indexer.Torrent from the ptt result
// by accumulating the ptt result with the torrent data
// not all of them is required, some are optional
func ParseTorrent(ctx context.Context, title string) *ptt.Result {
	logger := infrastructure.FromContext(ctx)

	logger.Debug("filter::ParseTorrent", "torrent", title)

	return ptt.Parse(title)
}
