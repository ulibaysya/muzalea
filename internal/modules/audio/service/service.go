package service

import (
	"context"
	"io"
	"log/slog"

	"github.com/ulibaysya/muzalea/internal/modules/mux"
	"github.com/ulibaysya/muzalea/internal/pkg/compensate"
)

type WaitGroup interface {
	Go(f func())
}

//go:generate go run github.com/vektra/mockery/v3@v3.8.0 --name=MetaStorage
type MetaStorage interface {
	// Save saves audio metadata and returns its ID, compensation function and
	// a possible error.
	//
	// The compensation function removes the saved metadata from storage. The
	// compensation function is nil if err is non-nil.
	Save(ctx context.Context, name, author string, tags []string) (id int64, cf compensate.Func, err error)

	// SaveStream stores a stream for the audio with the provided id.
	//
	// SaveStream may return ErrNotFound if no audio with the provided ID is
	// found.
	//
	// SaveStream updates the stream field. The previous value is overwritten
	// if one is set.
	SaveStream(ctx context.Context, id int64, stream string) (err error)
}

type MuxService interface {
	// Prepare checks audio and returns mux.State which can be used for Mux.
	Prepare(ctx context.Context, audio io.Reader) (ms mux.State, err error)

	// Mux processes earlier prepared state and saves it to storage.
	Mux(ctx context.Context, ms mux.State) (files []mux.File, err error)
}

type BlobStorage interface {
	// Save saves the contents of the provided body with the provided filename.
	//
	// The compensation function removes saved file.
	Save(ctx context.Context, filename string, body io.Reader) (cf compensate.Func, err error)
}

type Service struct {
	logger   *slog.Logger
	wg       WaitGroup
	metaStrg MetaStorage
	muxSrvc  MuxService
	blobStrg BlobStorage
}

func New(logger *slog.Logger, wg WaitGroup, muxSrvc MuxService, metaStrg MetaStorage, blobStrg BlobStorage) Service {
	return Service{
		logger:   logger,
		wg:       wg,
		metaStrg: metaStrg,
		muxSrvc:  muxSrvc,
		blobStrg: blobStrg,
	}
}
