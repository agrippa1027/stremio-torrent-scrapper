# Filtering Jackett results for series (Torbox streaming)

## Context

Jackett returns a lot of junk for series searches: wrong shows (title search is fuzzy), CAM/TS
rips, fake small "complete season" packs, and results for the wrong season or episode. Today
almost none of it is filtered.

Two findings shape the design:

**1. Stremio will not pick the episode for you.** The addon SDK spec says `fileIdx` is the index
of the video file "within the torrent (from `infoHash`)", and "if fileIdx is not specified, the
largest file in the torrent will be selected". There is no filename matching anywhere in Stremio
or its streaming server. And because we stream from Torbox we return a `url`, not an `infoHash` —
so `fileIdx` is ignored entirely; that URL *is* the file. **Episode→file resolution is entirely
our job.** Every established addon (Torrentio, StremThru, Comet) does exactly this.

**2. Season packs are worth keeping.** They are often better encodes and more likely to be cached
on Torbox. We already have the machinery to play them (`selectEpisodeFile` in
`internal/torbox/torbox_handlers.go`). The current code throws them away instead.

Four concrete bugs today:

- `filterEpisodeResults` (`internal/stream/stream.go:131`) drops **every** season pack whenever any
  episode-specific result exists — packs have no `SxxEyy` marker, so `extractEpisodeMarker` fails
  and they are filtered out. Its `if len(filtered) == 0 { return results }` fallback then lets pure
  junk through when nothing matches.
- **The `NxN` regex misparses resolutions.** `(?i)(\d{1,2})x(\d{1,2})` (`stream.go:127`, and its
  twin at `torbox_handlers.go:115`) has no anchors, so it matches *inside* a resolution. Verified by
  running the actual regex list:

  ```
  Breaking.Bad.Season.5.1920x1080.COMPLETE  -> S=20 E=10   (matched "20x10")
  Some.Show.1280x720.S02.Complete           -> S=80 E=72   (matched "80x72")
  Breaking.Bad.S05E16.1920x1080.BluRay      -> S=05 E=16   (correct — regex #0 wins first)
  ```

  Note exactly when it bites: a title containing `SxxEyy` matches regex #0 first and is fine. It
  only fires on titles *without* an episode marker — i.e. **season packs** — which then get a
  garbage season number and are rejected as "wrong season". So packs are being dropped for two
  independent reasons.
- `matchEpisodeFile` (`internal/torbox/torbox_handlers.go:132`) requires
  `strings.Contains(name, query)` where `query` is the **full release name**
  (`Breaking.Bad.S05E16.1080p.BluRay.x264-GROUP`). That essentially never matches a file name, so it
  silently falls back to "largest video file" — i.e. **season packs currently play the wrong
  episode**.
- `best.ID == 0` (same file, lines 107 and 140) treats file index 0 as "not found", rejecting a
  legitimate first file.

Scope, per your call: **filtering logic only**. No build fixes, no rework of the Torbox call flow.
Because `internal/stream` does not compile right now, the new logic goes in **fresh packages that
build and test standalone today**, and gets wired in when your refactor lands.

## Approach

### 1. `internal/release/` — one parser, replacing three copies

`episodeRegexes`/`extractEpisodeMarker` (`stream.go:125`) and `episodePatterns`/`parseEpisodeMarker`
(`torbox_handlers.go:113`) are character-for-character identical. `normalizeSeriesTitle` and friends
are a third partial copy in `series-search.go`. Consolidate:

```go
package release

type Info struct {
    Title      string  // normalized
    Year       int
    Seasons    []int   // multiple for S01-S03
    Episodes   []int   // multiple for S01E01E02; empty => pack
    Complete   bool
    Resolution string  // "2160p", "1080p", ...
    Source     string  // "bluray", "web-dl", "hdtv", "cam", ...
    Codec      string
    HDR, Remux bool
    Group      string
}

func Parse(name string) Info
func ParseFile(path string) Info                     // basename-aware, tolerant of extension
func NormalizeTitle(raw string) string
func MatchesTitle(rawTitle, normalizedWanted string, aliases []string) bool

func IsVideoFile(name string) bool   // widen beyond today's 4 extensions
func IsSampleFile(name string) bool  // sample/trailer/extras path segments
func IsArchive(name string) bool     // .rar .zip .iso .7z
```

`ParseFile` matters for the playback fix in §4: debrid services flatten and re-prefix paths, so
`Season 1/S01E05.mkv` must still resolve.

**Move, don't rewrite:** `matchesSeriesTitle`, `tokensMatch`, `isSeasonOrEpisodeToken`,
`isAllowedSeriesSuffix`, `normalizeSeriesTitle` at `internal/jackett/series-search.go:85-158` are
fully implemented, correct, and completely dead (nothing calls them; the call site is commented out
at lines 41-44). Lift them into this package as `MatchesTitle`/`NormalizeTitle`. This is the single
highest-value filter we have — it kills the "wrong show entirely" results.

**Parser backing — your call, and the plan is agnostic to it.** Every call site goes through
`release.Parse`, so this is a one-file swap either way:

- *Recommended:* `github.com/MunifTanjim/go-ptt`. Used in production by StremThru, a Go Stremio
  addon doing this exact job. Handles multi-episode, `Season 2` word form, date-based and anime
  absolute numbering, and exposes a `Complete` flag that is exactly our pack detector. It also
  classifies `CAM`/`TeleSync`/`TeleCine`/`SCR`/`R5` as first-class `Quality` values, which turns the
  trash filter into one set lookup instead of a substring blacklist. Deletes ~100 lines of regex.

  On dependency cost, since that's the usual objection: the root `ptt` package imports only stdlib
  (`regexp`, `strconv`, `strings`, `time`). The heavy things in its `go.mod` — grpc, protobuf,
  urfave/cli — belong to its `cmd/` and `packages/` subdirectories, and Go 1.17+ module graph
  pruning keeps them out of your build. Direct deps go 4 → 5, indirect unchanged. Worth confirming
  with `go mod tidy && go mod graph` before committing to it. MIT licensed. It is v0.x, so pin the
  exact version.
- *Alternative:* keep the hand-rolled regexes, consolidated here, and add the missing patterns
  yourself. No new dependency; you own the long tail of release-naming edge cases — anchoring the
  `NxN` pattern, `S01E01-E10`, `S01.E01`, `S01-S05`, `Season 1 Complete`, `1of10`, date-based
  episodes, anime absolute numbering.

Either way, keep `MatchesTitle`'s token-window comparison rather than exact string equality —
`ptt`'s title extraction is heuristic and can mis-split on anime and foreign releases, so compare
tolerantly against its already-stripped title.

### 2. `internal/filter/` — classify, reject, rank

```go
package filter

type Kind int
const (
    KindReject Kind = iota
    KindEpisode     // exact SxxEyy match
    KindSeasonPack  // covers requested season, no episode marker
    KindSeriesPack  // multi-season / complete series covering requested season
)

type Candidate struct {
    Result indexer.Result
    Info   release.Info
    Kind   Kind
    Score  int
    Reason string // rejection reason, for logging
}

type Options struct {
    SeriesTitle     string // from TMDB, for title matching
    Type            string
    Season, Episode int
    EpisodeCount    int    // optional; enables pack size sanity
    Cfg             config.FilterConfig
}

func Apply(results []indexer.Result, opts Options) []Candidate
```

Order of operations inside `Apply`:

**a. Dedup by infohash.** There is no dedup anywhere in the codebase today, and multiple trackers
routinely return the same torrent. Normalize case, fall back to
`infoHashFromMagnet` (`stream.go:214`, move it here) when `InfoHash` is empty; on collision keep the
higher seeder count.

**b. Classify** against the requested season/episode:

| Parsed | Condition | Kind |
|---|---|---|
| seasons + episodes | season matches && episode ∈ parsed.Episodes | `KindEpisode` |
| seasons + episodes | otherwise | `KindReject` |
| seasons, no episodes | requested season ∈ parsed.Seasons | `KindSeasonPack` (or `KindSeriesPack` if >1 season or `Complete`) |
| seasons, no episodes | otherwise | `KindReject` |
| neither | `Complete` set | `KindSeriesPack` |
| neither | otherwise | `KindReject` (unparseable ⇒ junk) |

Note there is deliberately **no `if len(filtered) == 0 { return results }` escape hatch**. An empty
list is better than a list of wrong episodes.

**c. Hard rejects**, each recording a `Reason`:

- **Title mismatch** vs the TMDB series title (`release.MatchesTitle`).
- **Trash source**: cam, ts, telesync, telecine, hdcam, screener/scr, workprint, r5, hdtc. For
  series these are essentially always fake or mislabeled.
- **Adult/spam categories**: Jackett `Categories` containing the 6000 (XXX) range; and when
  categories are present, require overlap with 5000 (TV) for series / 2000 (Movies) for movies.
- **Non-video / sample**: archive or disk containers (`.rar`, `.zip`, `.iso`, `.7z`), and
  `sample`/`trailer`/`extras` path segments.
- **Size sanity** — the filter that catches the classic scam, a "Season 1 Complete" that is 300 MB:
  ```
  N = 1                              for KindEpisode
  N = episodeCount()                 for KindSeasonPack
  N = episodeCount() * len(Seasons)  for KindSeriesPack
  perEpisode := Size / N
  reject if perEpisode < MinPerEpisodeMB  (default 50 MB)     // fake / sample
  reject if perEpisode > MaxPerEpisodeMB  (default 20 GB)     // mislabeled
  ```
  `episodeCount()` falls back in this order: TMDB `episode_count` for the season → **`Result.Files`
  clamped to 1..60** → `DefaultEpisodeCount` (10). `Files` is already populated by Jackett
  (`indexer.go`, mapped at `jackett.go:143`) and is a much better pack-size proxy than a flat
  constant.
- **Unplayable**: no infohash and no magnet, or zero size.

**d. Score** (replaces `sortResultsByQuality`, `stream.go:171`):

- resolution 2160p 400 / 1080p 300 / 720p 200 / 480p 100 — keep the existing table from `qualityRank`
- source: remux +100 / bluray +80 / web-dl +70 / webrip +55 / hdtv +20 / dvdrip +10;
  codec x265|hevc|av1 +20 / x264 +10; proper|repack +10
- HDR/DV +15, **but only at 2160p** — HDR tags on 1080p releases are frequently bogus
- **kind: `KindEpisode` +50, `KindSeasonPack` +25, `KindSeriesPack` +0** — prefer episode-specific
  (fewer ways to go wrong) but keep packs competitive, so a 1080p pack still beats a 720p single
- **seeders: `min(seeds,100)/10`, i.e. 0–10 points — a weak tiebreak, not a gate.** Rationale, and
  this is the direct consequence of streaming from Torbox rather than the torrent: for a torrent
  already cached on Torbox, seeders are irrelevant. They only matter if Torbox has to actually fetch
  it. A hard seeder floor would discard perfectly good cached results. `MinSeeders` exists as a
  config knob but **defaults to 0 (off)**; it becomes worth enabling only once we know cached status
  at filter time (see Future).
- freeleech (`DownloadVolumeFactor == 0`) +5

Sort descending, `sort.SliceStable`.

### 3. Config knobs

`internal/config/config.go` has zero filtering knobs today — everything is hardcoded. Add:

```go
type FilterConfig struct {
    MaxResults          int      `yaml:"maxresults"`          // 30
    MinSeeders          int      `yaml:"minseeders"`          // 0 — off; see rationale above
    MinPerEpisodeMB     int64    `yaml:"minperepisodemb"`     // 50
    MaxPerEpisodeMB     int64    `yaml:"maxperepisodemb"`     // 20000
    DefaultEpisodeCount int      `yaml:"defaultepisodecount"` // 10
    AllowSeasonPacks    *bool    `yaml:"allowseasonpacks"`    // default true
    AllowSeriesPacks    *bool    `yaml:"allowseriespacks"`    // default true
    RequireTitleMatch   *bool    `yaml:"requiretitlematch"`   // default true
    AllowTrashQuality   bool     `yaml:"allowtrashquality"`   // zero value false = reject ✓
    AllowAdult          bool     `yaml:"allowadult"`          // zero value false = reject ✓
    TrashQualities      []string `yaml:"trashqualities"`      // cam, ts, telesync, ...
    MinResolution       string   `yaml:"minresolution"`       // "" | 480p | 720p | 1080p
}
```

Add a `Filter FilterConfig \`yaml:"filter"\`` field to `Config`. Three things to get right:

1. **The `*bool` fields are deliberate.** A plain `bool` whose sane default is `true` cannot be
   distinguished from an explicitly-configured `false` — both arrive as the zero value. The flags
   that should default to *false* are instead named positively (`AllowTrashQuality`, `AllowAdult`)
   so their zero value is already the behaviour we want, and stay plain `bool`.
2. **Add `func (c *Config) ApplyDefaults()` and call it at the end of `LoadConfigFile`
   (`internal/config/loader.go`).** Without it, a `config.yml` with no `filter:` block yields
   `MinPerEpisodeMB = 0` and `MaxPerEpisodeMB = 0`, and the size-sanity check then **rejects every
   result**. The current `config.yml` has no such block, so this is a live footgun, not a
   hypothetical one.
3. Existing structs mix `json:` tags with a `yaml.Unmarshal` loader (`loader.go:17`) and only work
   by accident, via yaml.v3's lowercasing fallback — note that `baseUrl` in YAML would *not* bind
   but `baseurl` does. Use explicit lowercase `yaml:` tags on everything new; worth fixing
   `TorboxConfig`/`ManifestConfig` in the same pass.

### 4. Required companion fix: pack playback

Keeping season packs is pointless while they play the wrong episode. In
`internal/torbox/torbox_handlers.go`:

- `matchEpisodeFile`: drop the `strings.Contains(name, query)` condition entirely and match with
  `release.Parse(file.Name)` on season+episode. Match on the **basename** — debrid services flatten
  and re-prefix paths, so `Season 1/S01E05.mkv` must still match.
- Replace the `best.ID == 0` sentinel with an explicit `found bool` in both `matchEpisodeFile` and
  `selectEpisodeFile`.
- Widen `isVideoFile` beyond `.mkv/.mp4/.avi/.mov` — add at least `.m4v`, `.ts`, `.webm`, `.flv`,
  `.mpg`, `.mpeg`, `.wmv`.

Flagging this honestly as playback rather than filtering, but it is the change that makes "keep
season packs" correct rather than actively harmful.

### 5. Wiring (do when the refactor compiles)

- `internal/stream/stream.go:44-45` — replace the `filterEpisodeResults` + `sortResultsByQuality`
  pair with a single `filter.Apply(...)`, then delete `filterEpisodeResults`,
  `extractEpisodeMarker`, `sortResultsByQuality`, `qualityRank`, `episodeRegexes`. Note this sits
  **before** the serial `CheckCached` loop at line 60, so filtering here also cuts real latency.
- `internal/jackett/series-search.go:41-44` — delete the commented block and the now-migrated
  `filterSeriesResults` family.
- `internal/torbox/torbox_handlers.go:113-164` — delete `episodePatterns`/`parseEpisodeMarker`, use
  `release.Parse`.

## Files

| File | Change |
|---|---|
| `internal/release/release.go` | **new** — `Parse`, `NormalizeTitle`, `MatchesTitle` (latter two lifted from `series-search.go:85-158`) |
| `internal/filter/filter.go` | **new** — `Apply`, `Kind`, `Candidate`, dedup, classify, reject, score |
| `internal/filter/filter_test.go` | **new** — table-driven |
| `internal/release/release_test.go` | **new** — table-driven |
| `internal/config/config.go` | add `FilterConfig` + `Config.Filter` |
| `internal/torbox/torbox_handlers.go` | fix `matchEpisodeFile`, `best.ID == 0`, `isVideoFile` |
| `internal/stream/stream.go` | wire in `filter.Apply`, delete superseded helpers *(deferred — does not compile)* |
| `internal/jackett/series-search.go` | delete migrated + commented-out code *(deferred)* |

## Verification

`internal/release` and `internal/filter` have no dependency on the broken packages, so they build
and test on their own **today**:

```
go test ./internal/release/... ./internal/filter/...
```

Table-driven cases, using real Jackett titles for `tt0903747:5:16` (Breaking Bad S05E16):

- correct episode → `KindEpisode`, kept
- `S05E15`, `S04E16` → rejected (wrong episode / wrong season)
- `Breaking.Bad.S05.1080p.BluRay.x264` → `KindSeasonPack`, kept
- `Breaking.Bad.Complete.Series.1080p` → `KindSeriesPack`, kept
- `Breaking.Bad.S01.COMPLETE.720p` → rejected (season 1 ≠ 5)
- `Breaking.Bad.S05.COMPLETE.480p` at 300 MB → rejected (size sanity)
- `Breaking.Bad.S05E16.CAM` → rejected (trash source)
- `Better.Call.Saul.S05E16.1080p` → rejected (title mismatch)
- `S05E15E16` multi-episode → `KindEpisode`, kept
- same infohash from two trackers → deduped, higher seeders retained
- a 1080p season pack must outrank a 720p single episode
- **regression for the `NxN` bug**: `Breaking.Bad.Season.5.1920x1080.COMPLETE` → `KindSeasonPack`
  season 5, **not** S20E10; `Some.Show.1280x720.S02.Complete` → season 2, not S80E72
- `Breaking.Bad.S05.COMPLETE` with `Files=13` and no TMDB hint → per-episode size computed
  against 13, not the flat fallback
- `ApplyDefaults` on a config with no `filter:` block → results survive (guards the
  "every result rejected because MaxPerEpisodeMB is 0" footgun)
- **no results match → empty slice**, not the unfiltered input

Then, once the refactor compiles, end-to-end against a real series: hit
`/stream/series/tt0903747:5:16.json` and confirm the list contains no wrong-show/wrong-episode
entries, that season packs appear, and that selecting a pack plays episode 16 rather than the
longest file in the season.

## Future (explicitly out of scope now)

- **Batch the cache check.** Torbox `checkcached` accepts comma-separated hashes and a
  `list_files=true` param returning each torrent's file list. That collapses the 30 serial
  `CheckCached` calls into one **and** makes season-pack content verification free — you can confirm
  the pack actually contains the episode before offering it, and carry the matched filename into the
  stream URL. `client.go:201` currently hardcodes `list_files=false`. A cached pack whose file list
  demonstrably lacks the episode is the strongest reject signal available, and it costs nothing
  extra. Worth validating early: confirm that the `id` values from `checkcached?list_files=true`
  are usable as `requestdl`'s `file_id` after `CreateTorrent`. If they are, matching on ID is a free
  fast path; if not, the filename match in §4 carries it — which is why both should be carried
  rather than the ID alone.
- **Deferred resolution.** Today `buildTorboxStreamLinks` eagerly calls
  `CreateTorrent`+`GetTorrent`+`RequestDownload` per cached result at search time — up to ~120
  sequential calls, it adds every torrent to your account just to render a list, and Torbox links
  expire in 3 hours. Established addons return a URL to their own resolver and resolve on play. The
  code half-does this already, but `GET /torbox/stream/{hash}` is **not registered** in
  `internal/stremio/addon.go`, so the uncached path is currently a dead route.
- **Cache-aware seeder gate** — a hard floor for uncached results only, once (a) lands.
- **`behaviorHints`** — `bingeGroup` keyed per infohash so Stremio keeps using the same season pack
  for the next episode, plus `filename` and `videoSize`. No `BehaviorHints` field exists on `Stream`
  (`internal/types.go`) yet, and `FileIdx` is declared but never assigned.
