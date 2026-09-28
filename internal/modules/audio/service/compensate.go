package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ulibaysya/muzalea/internal/pkg/compensate"
	pkgslog "github.com/ulibaysya/muzalea/pkg/slog"
	"golang.org/x/sync/errgroup"
)

func (s Service) compensateMeta(ctx context.Context, cf compensate.Func) {
	if err := compensate.CallRetryably(cf, 3); err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "error while compensating audio metadata", pkgslog.Error("error", err))
		return
	}
	s.logger.LogAttrs(ctx, slog.LevelInfo, "successfully compensated audio metadata")
}

func (s Service) compensateBlobs(ctx context.Context, cs *compensate.Storage) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "starting audio blobs compensation")
	wg := errgroup.Group{}
	for _, unit := range cs.Take() {
		wg.Go(func() error {
			if err := compensate.CallRetryably(unit.Func, 3); err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "error while compensating audio blob",
					pkgslog.Error("error", err),
					slog.String("filename", unit.Meta[0].(string)),
				)
				return fmt.Errorf("compensating audio blob: %w", err)
			}
			s.logger.LogAttrs(ctx, slog.LevelInfo, "successfully compensated audio blob",
				slog.String("filename", unit.Meta[0].(string)),
			)
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return
	}
	s.logger.LogAttrs(ctx, slog.LevelInfo, "successfully compensated audio blobs")
}
