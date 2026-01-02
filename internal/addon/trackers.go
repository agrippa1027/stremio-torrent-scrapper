package addon

import (
	"os"
	"strings"
)

var defaultTrackers = []string{
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://tracker.openbittorrent.com:80/announce",
	"udp://tracker.leechers-paradise.org:6969/announce",
	"udp://tracker.coppersurfer.tk:6969/announce",
	"udp://tracker.internetwarriors.net:1337/announce",
}

func DefaultTrackers() []string {
	return append([]string(nil), defaultTrackers...)
}

func LoadTrackers(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	trackers := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		trackers = append(trackers, line)
	}

	return trackers, nil
}
