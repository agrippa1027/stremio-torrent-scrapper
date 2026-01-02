# Local Dataset Stremio Addon (Go)

A minimal Stremio addon that serves streams only from your own dataset and public torrent trackers.

## Run

```bash
go run ./cmd/addon
```

The addon listens on `:7000` by default.

- Manifest: `http://localhost:7000/manifest.json`
- Streams: `http://localhost:7000/stream/movie/tt0111161.json`
- Series: `http://localhost:7000/stream/series/tt0903747:1:1.json`

## Project layout

- `cmd/addon`: entrypoint
- `internal/addon`: addon logic and HTTP handlers
- `data`: local dataset and trackers list

## Dataset format

Edit `data/dataset.json` to match your content. Example:

```json
[
  {
    "type": "movie",
    "id": "tt0111161",
    "title": "The Shawshank Redemption",
    "infoHash": "0123456789ABCDEF0123456789ABCDEF01234567",
    "fileIdx": 0,
    "size": 2147483648,
    "seeds": 120,
    "resolution": "1080p"
  },
  {
    "type": "series",
    "id": "tt0903747",
    "title": "Breaking Bad S01E01",
    "infoHash": "89ABCDEF0123456789ABCDEF0123456789ABCDEF",
    "fileIdx": 0,
    "size": 734003200,
    "seeds": 42,
    "resolution": "720p",
    "episode": { "season": 1, "episode": 1 }
  }
]
```

## Trackers

Edit `data/trackers.txt` to control the public trackers used to build magnet URIs.

## Config

- `PORT`: server port (default `7000`)
- `DATASET_PATH`: dataset JSON path (default `data/dataset.json`)
- `TRACKERS_PATH`: trackers list path (default `data/trackers.txt`)
