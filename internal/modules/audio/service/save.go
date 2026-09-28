package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"

	"github.com/ulibaysya/muzalea/internal/modules/mux"
	"github.com/ulibaysya/muzalea/internal/pkg/compensate"
	pkgrand "github.com/ulibaysya/muzalea/pkg/rand"
	pkgslog "github.com/ulibaysya/muzalea/pkg/slog"
	"golang.org/x/sync/errgroup"
)

func (s Service) SaveAudio(ctx context.Context, name, author string, tags []string, audio io.Reader) (int64, error) {
	muxState, err := s.muxSrvc.Prepare(ctx, audio)
	if err != nil {
		return 0, fmt.Errorf("preparing audio for mux: %w", err)
	}
	defer func() {
		if err != nil {
			muxState.Close() //nolint:errcheck
		}
	}()

	audioID, metaCompensate, err := s.metaStrg.Save(ctx, name, author, tags)
	if err != nil {
		return 0, fmt.Errorf("saving audio meta: %w", err)
	}

	s.wg.Go(func() { s.muxAudio(ctx, muxState, audioID, metaCompensate) })

	return audioID, nil
}

func (s Service) muxAudio(ctx context.Context, muxState mux.State, audioID int64, metaComp compensate.Func) {
	defer muxState.Close() //nolint:errcheck

	muxFiles, err := s.muxSrvc.Mux(ctx, muxState)
	defer func() {
		if err != nil {
			s.compensateMeta(ctx, metaComp)
		}
	}()
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "error while muxing audio", pkgslog.Error("error", err))
		return
	}

	playlist, blobsCs, err := s.saveBlobs(ctx, muxFiles)
	defer func() {
		if err != nil {
			s.compensateBlobs(ctx, blobsCs)
		}
	}()
	if err != nil {
		return
	}

	s.logger.LogAttrs(ctx, slog.LevelDebug, "saving stream", slog.String("playlist", playlist), slog.Int64("id", audioID))
	if err = s.metaStrg.SaveStream(ctx, audioID, playlist); err != nil {
		return
	}
}

func (s Service) saveBlobs(ctx context.Context, muxFiles []mux.File) (string, *compensate.Storage, error) {
	prefix := pkgrand.Base64URL(32)
	wg, ctx := errgroup.WithContext(ctx)
	// wg := errgroup.Group{}
	compStrg := compensate.NewStorage()
	playlistFilename := ""
	for _, muxFile := range muxFiles {
		filename := path.Join(prefix, muxFile.Name)
		if muxFile.Playlist {
			playlistFilename = filename
		}
		// TODO use errgroup with context
		wg.Go(func() error {
			cf, err := s.blobStrg.Save(ctx, filename, muxFile.Contents)
			if err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "error while saving to blog storage",
					pkgslog.Error("error", err),
					slog.String("filename", filename),
				)
				return fmt.Errorf("saving to blog storage: %w", err)
			}
			compStrg.Append(compensate.StorageUnit{
				Func: cf,
				Meta: []any{filename},
			})
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return "", compStrg, err
	}
	return playlistFilename, compStrg, nil
}
