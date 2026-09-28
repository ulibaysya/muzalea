package service

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/ulibaysya/muzalea/internal/modules/mux"
	"github.com/ulibaysya/muzalea/internal/pkg/compensate"
)

func TestSaveAudio_PrepareFail(t *testing.T) {
	ctx := t.Context()
	name := "name"
	author := "author"
	tags := []string{"tag1", "tag2"}
	audio := strings.NewReader("audio")

	someErr := errors.New("some error")

	muxSrvc := NewMockMuxService(t)
	muxSrvc.
		On("Prepare", ctx, audio).
		Return(nil, someErr).
		Once()

	srvc := New(slog.Default(), NewMockWaitGroup(t), muxSrvc, NewMockMetaStorage(t), NewMockBlobStorage(t))
	id, err := srvc.SaveAudio(ctx, name, author, tags, audio)

	if id != 0 {
		t.Fatalf("expected id 0, got %v", id)
	}

	if !errors.Is(err, someErr) {
		t.Fatalf("expected err %v, got %v", someErr, err)
	}
}

func TestSaveAudio_MetaSaveFail(t *testing.T) {
	ctx := t.Context()
	name := "name"
	author := "author"
	tags := []string{"tag1", "tag2"}
	audio := strings.NewReader("audio")
	someErr := errors.New("some error")

	muxState := NewMockState(t)
	muxState.
		On("Close").
		Return(nil).
		Once()

	muxSrvc := NewMockMuxService(t)
	muxSrvc.
		On("Prepare", ctx, audio).
		Return(muxState, nil).
		Once()

	metaStrg := NewMockMetaStorage(t)
	metaStrg.
		On("Save", ctx, name, author, tags).
		Return(int64(0), compensate.Func(nil), someErr).
		Once()

	srvc := New(slog.Default(), NewMockWaitGroup(t), muxSrvc, metaStrg, NewMockBlobStorage(t))
	id, err := srvc.SaveAudio(ctx, name, author, tags, audio)

	if id != 0 {
		t.Fatalf("expected id 0, got %v", id)
	}

	if !errors.Is(err, someErr) {
		t.Fatalf("expected err %v, got %v", someErr, err)
	}
}

func TestSaveAudio_AsyncMuxStarted(t *testing.T) {
	ctx := t.Context()
	name := "name"
	author := "author"
	tags := []string{"tag1", "tag2"}
	audio := strings.NewReader("audio")

	wg := NewMockWaitGroup(t)
	wg.
		On("Go", mock.Anything).
		Once()

	muxSrvc := NewMockMuxService(t)
	muxSrvc.
		On("Prepare", ctx, audio).
		Return(NewMockState(t), nil).
		Once()

	metaStrg := NewMockMetaStorage(t)
	metaStrg.
		On("Save", ctx, name, author, tags).
		// Compensation is expected not to be called when async mux is started.
		// A nil function is returned so the test panics if it is called.
		Return(int64(5), compensate.Func(nil), nil).
		Once()

	srvc := New(slog.Default(), wg, muxSrvc, metaStrg, NewMockBlobStorage(t))
	id, err := srvc.SaveAudio(ctx, name, author, tags, audio)

	if id != 5 {
		t.Fatalf("expected id 5, got %v", id)
	}

	if err != nil {
		t.Fatalf("expected err nil, got %v", err)
	}
}

func TestMuxAudio_FailMux(t *testing.T) {
	ctx := t.Context()
	someErr := errors.New("some error")

	muxState := NewMockState(t)
	muxState.
		On("Close").
		Return(nil).
		Once()

	muxSrvc := NewMockMuxService(t)
	muxSrvc.
		On("Mux", ctx, muxState).
		Return(nil, someErr).
		Once()

	srvc := New(slog.Default(), NewMockWaitGroup(t), muxSrvc, NewMockMetaStorage(t), NewMockBlobStorage(t))
	metaCompensationCalls := 0
	srvc.muxAudio(ctx, muxState, 5, func(ctx context.Context) error {
		metaCompensationCalls += 1
		return nil
	})

	if metaCompensationCalls != 1 {
		t.Fatalf("expected meta compensation call 1 time, got %v", metaCompensationCalls)
	}
}

func TestMuxAudio_FailSaveBlobs(t *testing.T) {
	ctx := t.Context()
	someErr := errors.New("some error")
	muxFiles := []mux.File{
		{
			Name:     "1",
			Contents: strings.NewReader("1"),
			Playlist: true,
		},
		{
			Name:     "2",
			Contents: strings.NewReader("2"),
		},
	}

	muxState := NewMockState(t)
	muxState.
		On("Close").
		Return(nil).
		Once()

	muxSrvc := NewMockMuxService(t)
	muxSrvc.
		On("Mux", ctx, muxState).
		Return(muxFiles, nil).
		Once()

	blobStrg := NewMockBlobStorage(t)
	blobCompensationCalls := 0
	blobStrg.
		On("Save", mock.Anything, mock.MatchedBy(func(filename string) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{43}/1$`).MatchString(filename)
		}), muxFiles[0].Contents).
		Return(compensate.Func(func(ctx context.Context) error {
			blobCompensationCalls += 1
			return nil
		}), nil).
		On("Save", mock.Anything, mock.MatchedBy(func(filename string) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{43}/2$`).MatchString(filename)
		}), muxFiles[1].Contents).
		// Compensation is expected to be nil when blob save fails. So it is
		// expected not to be called. Test panics if it is called.
		Return(compensate.Func(nil), someErr)

	srvc := New(slog.Default(), NewMockWaitGroup(t), muxSrvc, NewMockMetaStorage(t), blobStrg)
	metaCompensationCalls := 0
	srvc.muxAudio(ctx, muxState, 5, func(ctx context.Context) error {
		metaCompensationCalls += 1
		return nil
	})

	if blobCompensationCalls != 1 {
		t.Fatalf("expected blob compensation call 1 time, got %v", blobCompensationCalls)
	}

	if metaCompensationCalls != 1 {
		t.Fatalf("expected meta compensation call 1 time, got %v", metaCompensationCalls)
	}
}

func TestMuxAudio_FailSaveStream(t *testing.T) {
	ctx := t.Context()
	someErr := errors.New("some error")
	muxFiles := []mux.File{
		{
			Name:     "1",
			Contents: strings.NewReader("1"),
			Playlist: true,
		},
		{
			Name:     "2",
			Contents: strings.NewReader("2"),
		},
	}

	muxState := NewMockState(t)
	muxState.
		On("Close").
		Return(nil).
		Once()

	muxSrvc := NewMockMuxService(t)
	muxSrvc.
		On("Mux", ctx, muxState).
		Return(muxFiles, nil).
		Once()

	blobStrg := NewMockBlobStorage(t)
	firstBlobCompensationCalls := 0
	secondBlobCompensationCalls := 0
	blobStrg.
		On("Save", mock.Anything, mock.MatchedBy(func(filename string) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{43}/1$`).MatchString(filename)
		}), muxFiles[0].Contents).
		Return(compensate.Func(func(ctx context.Context) error {
			firstBlobCompensationCalls += 1
			return nil
		}), nil).
		On("Save", mock.Anything, mock.MatchedBy(func(filename string) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{43}/2$`).MatchString(filename)
		}), muxFiles[1].Contents).
		Return(compensate.Func(func(ctx context.Context) error {
			secondBlobCompensationCalls += 1
			return nil
		}), nil)

	metaStrg := NewMockMetaStorage(t)
	metaStrg.
		On("SaveStream", ctx, int64(5), mock.MatchedBy(func(filename string) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{43}/1$`).MatchString(filename)
		})).
		Return(someErr).
		Once()

	srvc := New(slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug})), NewMockWaitGroup(t), muxSrvc, metaStrg, blobStrg)
	metaCompensationCalls := 0
	srvc.muxAudio(ctx, muxState, 5, func(ctx context.Context) error {
		metaCompensationCalls += 1
		return nil
	})

	if firstBlobCompensationCalls != 1 {
		t.Fatalf("expected first blob compensation call 1 time, got %v", firstBlobCompensationCalls)
	}

	if secondBlobCompensationCalls != 1 {
		t.Fatalf("expected second blob compensation call 1 time, got %v", secondBlobCompensationCalls)
	}

	if metaCompensationCalls != 1 {
		t.Fatalf("expected meta compensation call 1 time, got %v", metaCompensationCalls)
	}
}

func TestMuxAudio_Success(t *testing.T) {
	ctx := t.Context()
	muxFiles := []mux.File{
		{
			Name:     "1",
			Contents: strings.NewReader("1"),
			Playlist: true,
		},
		{
			Name:     "2",
			Contents: strings.NewReader("2"),
		},
	}

	muxState := NewMockState(t)
	muxState.
		On("Close").
		Return(nil).
		Once()

	muxSrvc := NewMockMuxService(t)
	muxSrvc.
		On("Mux", ctx, muxState).
		Return(muxFiles, nil).
		Once()

	blobStrg := NewMockBlobStorage(t)
	firstBlobCompensationCalls := 0
	secondBlobCompensationCalls := 0
	blobStrg.
		On("Save", mock.Anything, mock.MatchedBy(func(filename string) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{43}/1$`).MatchString(filename)
		}), muxFiles[0].Contents).
		Return(compensate.Func(func(ctx context.Context) error {
			firstBlobCompensationCalls += 1
			return nil
		}), nil).
		On("Save", mock.Anything, mock.MatchedBy(func(filename string) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{43}/2$`).MatchString(filename)
		}), muxFiles[1].Contents).
		Return(compensate.Func(func(ctx context.Context) error {
			secondBlobCompensationCalls += 1
			return nil
		}), nil)

	metaStrg := NewMockMetaStorage(t)
	metaStrg.
		On("SaveStream", ctx, int64(5), mock.MatchedBy(func(filename string) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{43}/1$`).MatchString(filename)
		})).
		Return(nil).
		Once()

	srvc := New(slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug})), NewMockWaitGroup(t), muxSrvc, metaStrg, blobStrg)
	metaCompensationCalls := 0
	srvc.muxAudio(ctx, muxState, 5, func(ctx context.Context) error {
		metaCompensationCalls += 1
		return nil
	})

	if firstBlobCompensationCalls != 0 {
		t.Fatalf("expected first blob compensation call 1 time, got %v", firstBlobCompensationCalls)
	}

	if secondBlobCompensationCalls != 0 {
		t.Fatalf("expected second blob compensation call 1 time, got %v", secondBlobCompensationCalls)
	}

	if metaCompensationCalls != 0 {
		t.Fatalf("expected meta compensation call 1 time, got %v", metaCompensationCalls)
	}
}
