package service

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	pkgslog "github.com/ulibaysya/muzalea/pkg/slog"
)

func (s Service) Save(ctx context.Context, filename string, body io.Reader) {
	s.wg.Go(func() { s.save(ctx, filename, body) })
}

func (s Service) save(ctx context.Context, filename string, body io.Reader) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "reading all body", pkgslog.Error("error", err))
		return
	}
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: s.s3Bucket,
		Key:    new(filepath.Join(s.prefix, filename)),
		// Body:   body,
		Body: bytes.NewReader(bodyBytes),
	})
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "putting object to s3",
			slog.String("filename", filename),
			pkgslog.Error("error", err),
		)
	}
	s.logger.LogAttrs(ctx, slog.LevelInfo, "successfully put", slog.String("filename", filename))
}
