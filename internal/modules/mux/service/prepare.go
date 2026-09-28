package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/asticode/go-astiav"
	"github.com/ulibaysya/muzalea/internal/modules/mux"
)

func (s Service) Prepare(ctx context.Context, audio io.Reader) (mux.State, error) {
	inFmtCtx := astiav.AllocFormatContext()
	if inFmtCtx == nil {
		return nil, fmt.Errorf("nil input format context")
	}

	inIOCtx, err := astiav.AllocIOContext(4096, false, audio.Read, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("allocating input io context: %w", err)
	}
	if inIOCtx == nil {
		return nil, fmt.Errorf("nil input io context")
	}

	inFmtCtx.SetPb(inIOCtx)

	if err = inFmtCtx.OpenInput("", nil, nil); err != nil {
		return nil, fmt.Errorf("opening input: %w", err)
	}

	if err = inFmtCtx.FindStreamInfo(nil); err != nil {
		return nil, fmt.Errorf("finding stream info: %w", err)
	}

	inStreams := inFmtCtx.Streams()
	if inStreamsLen := len(inStreams); inStreamsLen != 1 {
		return nil, fmt.Errorf("not 1 input stream: %v", inStreamsLen)
	}
	inStream := inStreams[0]
	inStreamCodec := inStream.CodecParameters()
	if inStreamMt := inStreamCodec.MediaType(); inStreamMt != astiav.MediaTypeAudio {
		return nil, fmt.Errorf("input stream media-type is not an audio: %v", inStreamMt)
	}

	s.logger.LogAttrs(ctx, slog.LevelInfo, "prepared audio",
		slog.String("fileFormat", inFmtCtx.InputFormat().Name()),
		slog.String("audioCodec", inStreamCodec.CodecID().Name()),
		slog.Duration("audioDuration", time.Duration(inStream.Duration())),
	)

	return astiavState{
		inFmtCtx:      inFmtCtx,
		inIOCtx:       inIOCtx,
		inAudioStream: inStream,
	}, nil
}

type astiavState struct {
	inFmtCtx      *astiav.FormatContext
	inIOCtx       *astiav.IOContext
	inAudioStream *astiav.Stream
}

func (s astiavState) Close() error {
	if s.inFmtCtx != nil {
		s.inFmtCtx.CloseInput()
	}
	if s.inIOCtx != nil {
		s.inIOCtx.Free()
	}
	s.inFmtCtx.Free()

	return nil
}

func (s astiavState) State() {}
