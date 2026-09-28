package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/asticode/go-astiav"
	"github.com/ulibaysya/muzalea/internal/modules/blob"
	"github.com/ulibaysya/muzalea/internal/modules/mux"
)

func (s Service) Mux(ctx context.Context, ms mux.State, bs blob.Storage) error {
	state, ok := ms.(astiavState)
	if !ok {
		return fmt.Errorf("invalid state")
	}
	defer state.Close() //nolint:errcheck

	outFmtCtx, err := astiav.AllocOutputFormatContext(nil, "", "playlist.m3u8")
	if err != nil {
		return fmt.Errorf("allocating output format context: %w", err)
	}
	if outFmtCtx == nil {
		return fmt.Errorf("nil output format context: %w", err)
	}
	defer outFmtCtx.Free()

	options := []struct {
		key string
		val string
	}{
		{"hls_segment_type", "fmp4"},
		{"hls_time", "4"},
		{"hls_list_size", "0"},
		{"hls_segment_filename", "seg%04d.m4s"},
		{"hls_fmp4_init_filename", "init.mp4"},
		{"hls_playlist_type", "vod"},
	}
	optionsDictionary := astiav.NewDictionary()
	defer optionsDictionary.Free()
	for _, i := range options {
		if err = optionsDictionary.Set(i.key, i.val, 0); err != nil {
			return fmt.Errorf("setting %v option with %v: %w", i.key, i.val, err)
		}
	}

	mu := sync.Mutex{}
	pipeWriters := map[*astiav.IOContext]*io.PipeWriter{}
	outFmtCtx.SetIOOpener(
		func(filename string, flags astiav.IOContextFlags, d *astiav.Dictionary) (*astiav.IOContext, error) {
			pr, pw := io.Pipe()

			ioContext, err := astiav.AllocIOContext(4096, true, nil, nil, pw.Write)
			if err != nil {
				return nil, err
			}

			mu.Lock()
			pipeWriters[ioContext] = pw
			mu.Unlock()

			bs.Save(ctx, filename, pr)
			s.logger.Info("saving", "filename", filename)

			return ioContext, nil
		},
		func(ioContext *astiav.IOContext) error {
			defer ioContext.Free()

			ioContext.Flush()

			mu.Lock()
			pw, ok := pipeWriters[ioContext]
			delete(pipeWriters, ioContext)
			mu.Unlock()
			if !ok {
				s.logger.LogAttrs(ctx, slog.LevelError, "can't find pipe writer in io close")
				return astiav.ErrBug
			}
			if err := pw.Close(); err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "closing pipe writer", slog.Any("error", err))
				return astiav.ErrEio
			}

			return nil
		},
	)

	outAudioStream := outFmtCtx.NewStream(nil)
	if err = state.inAudioStream.CodecParameters().Copy(outAudioStream.CodecParameters()); err != nil {
		return fmt.Errorf("copying codes parameters: %w", err)
	}
	outAudioStream.CodecParameters().SetCodecTag(0)

	if err = outFmtCtx.WriteHeader(optionsDictionary); err != nil {
		return fmt.Errorf("main: writing header failed: %w", err)
	}

	pkt := astiav.AllocPacket()
	defer pkt.Free()
	for {
		if stop := func() bool {
			if err = state.inFmtCtx.ReadFrame(pkt); err != nil {
				if !errors.Is(err, astiav.ErrEof) {
					s.logger.Info(fmt.Errorf("main: reading frame: %w", err).Error())
				}
				return true
			}

			defer pkt.Unref()

			pkt.SetStreamIndex(outAudioStream.Index())
			pkt.RescaleTs(state.inAudioStream.TimeBase(), outAudioStream.TimeBase())
			pkt.SetPos(-1)

			if err = outFmtCtx.WriteInterleavedFrame(pkt); err != nil {
				s.logger.Info(fmt.Errorf("main: writing interleaved frame: %w", err).Error())
				return true
			}
			return false
		}(); stop {
			break
		}
	}

	if err = outFmtCtx.WriteTrailer(); err != nil {
		return fmt.Errorf("main: writing trailer failed: %w", err)
	}

	s.logger.Info("success")

	return nil
}
