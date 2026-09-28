package traceid

import (
	"context"
	"log/slog"
)

func NewSlogHandler(next slog.Handler) slog.Handler {
	return slogHandler{
		next: next,
	}
}

type slogHandler struct {
	next slog.Handler
}

func (h slogHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.next.Enabled(ctx, lvl)
}

func (h slogHandler) Handle(ctx context.Context, rec slog.Record) error {
	id, ok := ExtractFromContext(ctx)
	if !ok {
		return h.next.Handle(ctx, rec)
	}
	rec.AddAttrs(slog.String("traceID", id))
	return h.next.Handle(ctx, rec)
}

func (h slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.next = h.next.WithAttrs(attrs)
	return h
}

func (h slogHandler) WithGroup(name string) slog.Handler {
	h.next = h.next.WithGroup(name)
	return h
}
