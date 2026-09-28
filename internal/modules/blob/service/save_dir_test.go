package service
//
// import (
// 	"log/slog"
// 	"testing"
//
// 	"github.com/aws/aws-sdk-go-v2/credentials"
// 	"github.com/aws/aws-sdk-go-v2/service/s3"
//
// 	// "github.com/google/uuid"
// 	"github.com/ulibaysya/muzalea/pkg/rand"
// )
//
// func TestSaveDir(t *testing.T) {
// 	client := s3.New(s3.Options{
// 		BaseEndpoint: new("http://127.0.0.1:9000"),
// 		Credentials:  credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", ""),
// 		Region:       "us-east-1",
// 	})
//
// 	// err := New(client).SaveDir(t.Context(), uuid.NewString(), "/home/ulibaysya/last/muzalea/tmp/hls/ffmpeg/output")
// 	err := New(slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug})), client, "audio").SaveDir(t.Context(), rand.Base64URL(32), "/home/ulibaysya/last/muzalea/tmp/hls/ffmpeg/output")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
