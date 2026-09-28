package service
//
// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"log/slog"
// 	"os"
// 	"path/filepath"
//
// 	"github.com/aws/aws-sdk-go-v2/service/s3"
// 	pkgslog "github.com/ulibaysya/muzalea/pkg/slog"
// )
//
// func (s Service) SaveDir(ctx context.Context, identifier string, dir string) (err error) {
// 	dirEntries, err := os.ReadDir(dir)
// 	if err != nil {
// 		return fmt.Errorf("reading dir entries: %w", err)
// 	}
// 	for _, i := range dirEntries {
// 		if !i.Type().IsRegular() {
// 			continue
// 		}
//
// 		entryName := i.Name()
// 		entryRealPath := filepath.Join(dir, entryName)
//
// 		var entryFile io.ReadCloser
// 		entryFile, err = os.Open(entryRealPath)
// 		if err != nil {
// 			return fmt.Errorf("opening file: %w", err)
// 		}
// 		defer func() {
// 			if closingErr := entryFile.Close(); closingErr != nil {
// 				s.logger.LogAttrs(ctx, slog.LevelWarn, "error when closing file",
// 					pkgslog.Error("error", closingErr),
// 					slog.String("realpath", entryRealPath),
// 				)
// 			}
// 		}()
//
// 		_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
// 			Bucket: s.s3Bucket,
// 			Key:    new(filepath.Join(identifier, entryName)),
// 			Body:   entryFile,
// 		})
// 		if err != nil {
// 			return fmt.Errorf("uploading %v: %w", entryName, err)
// 		}
// 	}
//
// 	return nil
// }
