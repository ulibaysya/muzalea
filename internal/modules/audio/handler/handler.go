package handler

import (
	"context"
	"io"
	"log/slog"
)

type Service interface {
	// SaveAudio saves name, author and tags in some meta storage.
	//
	// Then SaveAudio checks audio in some way, starts async processing
	// (probably remux) and then returns audio ID. If any of this operation
	// fails (checking and async processing start) SaveAudio deletes meta
	// (name, author and tags) from storage.
	//
	// After processing SaveAudio saves processed audio in some audio storage.
	SaveAudio(ctx context.Context, name, author string, tags []string, audio io.Reader) (int64, error)
}

type Handler struct {
	logger  *slog.Logger
	service Service
}

func New(logger *slog.Logger, service Service) Handler {
	return Handler{
		logger:  logger,
		service: service,
	}
}
