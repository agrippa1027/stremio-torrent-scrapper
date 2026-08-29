package filter

import (
	"context"

	ptt "github.com/MunifTanjim/go-ptt"
)

// This function should produce the torrent result from the parsed title.
// by accumulating the ptt result with the torrent data
// not all of them is required, some are optional
func ParseTorrent(ctx context.Context, title string) *ptt.Result {
	return ptt.Parse(title)
}
