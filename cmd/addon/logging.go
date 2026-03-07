package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
