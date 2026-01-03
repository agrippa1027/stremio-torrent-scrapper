# Local Dataset Stremio Addon (Go)

A minimal Stremio addon that returns streams from Jackett based on a title search.

## Run

```bash
go run ./cmd/addon
```

The addon listens on `:7000` by default.

- Manifest: `http://localhost:7000/manifest.json`
- Streams: `http://localhost:7000/stream/movie/The%20Matrix.json`
- Series: `http://localhost:7000/stream/series/Breaking%20Bad:1:1.json`

## Requirements

- Go 1.24+ (required by `go-jackett`)
- A running Jackett instance with configured indexers

## Project layout

- `cmd/addon`: entrypoint
- `internal/addon`: addon logic and HTTP handlers
- `internal/indexer`: indexer-specific scrapers

## Trackers

This addon uses a built-in list of public trackers to build magnet URIs.

## How it works

The stream endpoint treats the `{id}` portion of the path as a URL-encoded title
query. For series requests, it appends `SxxEyy` to the query. The addon then
queries Jackett (which aggregates public indexers), gathers the top results, and
returns Stremio streams backed by magnet links.

## Config

You can configure via `config.yml` or environment variables (env overrides file).

Example `config.yml`:

```yaml
port: "7000"
jacketturl: "http://localhost:9117"
jackettapikey: "REPLACE_ME"
```

Environment variables:

- `PORT`: server port (default `7000`)
- `JACKETT_API_URL`: Jackett base URL (e.g., `http://localhost:9117`)
- `JACKETT_API_KEY`: Jackett API key (from the Jackett UI)

## Logging

Logs are emitted as JSON to stdout, including request info and per-indexer
success/failure counts.
