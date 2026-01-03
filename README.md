# Local Dataset Stremio Addon (Go)

A minimal Stremio addon that returns streams from public torrent indexers based on a title search.

## Run

```bash
go run ./cmd/addon
```

The addon listens on `:7000` by default.

- Manifest: `http://localhost:7000/manifest.json`
- Streams: `http://localhost:7000/stream/movie/The%20Matrix.json`
- Series: `http://localhost:7000/stream/series/Breaking%20Bad:1:1.json`

## Project layout

- `cmd/addon`: entrypoint
- `internal/addon`: addon logic and HTTP handlers
- `internal/indexer`: indexer-specific scrapers

## Trackers

This addon uses a built-in list of public trackers to build magnet URIs.

## How it works

The stream endpoint treats the `{id}` portion of the path as a URL-encoded title
query. For series requests, it appends `SxxEyy` to the query. The addon then
searches public indexers (currently YTS and 1337x), gathers the top results, and
returns Stremio streams backed by magnet links.

## Config

- `PORT`: server port (default `7000`)
