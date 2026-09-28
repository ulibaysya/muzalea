package service

import (
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Service struct {
	prefix   string
	logger   *slog.Logger
	wg       *sync.WaitGroup
	s3Client *s3.Client
	s3Bucket *string
}

func New(logger *slog.Logger, prefix string, wg *sync.WaitGroup, s3Client *s3.Client, s3Bucket string) Service {
	return Service{
		logger:   logger,
		prefix:   prefix,
		wg:       wg,
		s3Client: s3Client,
		s3Bucket: &s3Bucket,
	}
}
