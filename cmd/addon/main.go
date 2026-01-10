package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"scrapper/internal/addon"

	"gopkg.in/yaml.v3"
)

func main() {
	cfg := addon.DefaultConfig()
	loadConfigFile("config.yml", &cfg)
	applyEnvOverrides(&cfg)

	port := cfg.Port

	logger := slog.New(newColorHandler(os.Stdout, slog.LevelDebug))
	slog.SetDefault(logger)

	app, err := addon.New(cfg, logger)
	if err != nil {
		log.Fatalf("init addon: %v", err)
	}

	mux := http.NewServeMux()
	app.Register(mux)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}

func loadConfigFile(path string, cfg *addon.Config) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	_ = yaml.Unmarshal(data, cfg)
}

func applyEnvOverrides(cfg *addon.Config) {
	if v := os.Getenv("JACKETT_API_URL"); v != "" {
		cfg.JackettURL = v
	}
	if v := os.Getenv("JACKETT_API_KEY"); v != "" {
		cfg.JackettAPIKey = v
	}
	if v := os.Getenv("TMDB_API_KEY"); v != "" {
		cfg.TMDBAPIKey = v
	}
	if v := os.Getenv("TORBOX_API_KEY"); v != "" {
		cfg.TorboxAPIKey = v
	}
	if v := os.Getenv("TORBOX_BASE_URL"); v != "" {
		cfg.TorboxBaseURL = v
	}
	if v := os.Getenv("TORBOX_CACHED_ONLY"); v != "" {
		cfg.TorboxCachedOnly = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}
	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}
}

type colorHandler struct {
	w      io.Writer
	min    slog.Level
	attrs  []slog.Attr
	groups []string
	mu     sync.Mutex
}

func newColorHandler(w io.Writer, min slog.Level) *colorHandler {
	return &colorHandler{w: w, min: min}
}

func (h *colorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.min
}

func (h *colorHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	if !r.Time.IsZero() {
		b.WriteString("time=")
		b.WriteString(r.Time.Format(time.RFC3339Nano))
		b.WriteByte(' ')
	}
	b.WriteString("level=")
	b.WriteString(colorLevel(r.Level))
	b.WriteByte(' ')
	b.WriteString("msg=")
	appendValue(&b, r.Message)

	for _, attr := range h.attrs {
		appendAttr(&b, h.groups, attr)
	}
	r.Attrs(func(a slog.Attr) bool {
		appendAttr(&b, h.groups, a)
		return true
	})

	b.WriteByte('\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, b.String())
	return err
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &clone
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	clone := *h
	if name != "" {
		clone.groups = append(append([]string{}, h.groups...), name)
	}
	return &clone
}

func appendAttr(b *strings.Builder, groups []string, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}
	b.WriteByte(' ')
	if len(groups) > 0 {
		b.WriteString(strings.Join(groups, "."))
		b.WriteByte('.')
	}
	b.WriteString(attr.Key)
	b.WriteByte('=')
	appendValue(b, attr.Value.Any())
}

func appendValue(b *strings.Builder, v any) {
	text := fmt.Sprint(v)
	if needsQuoting(text) {
		b.WriteString(strconv.Quote(text))
		return
	}
	b.WriteString(text)
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' || r == '\\' {
			return true
		}
	}
	return false
}

func colorLevel(level slog.Level) string {
	const (
		reset  = "\033[0m"
		red    = "\033[31m"
		yellow = "\033[33m"
		green  = "\033[32m"
		blue   = "\033[34m"
	)

	color := blue
	switch {
	case level >= slog.LevelError:
		color = red
	case level >= slog.LevelWarn:
		color = yellow
	case level >= slog.LevelInfo:
		color = green
	}

	return color + level.String() + reset
}
