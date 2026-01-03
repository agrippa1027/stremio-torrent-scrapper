package addon

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
