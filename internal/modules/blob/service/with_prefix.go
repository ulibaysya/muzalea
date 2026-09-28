package service

import (
	"path/filepath"

	"github.com/ulibaysya/muzalea/internal/modules/blob"
)

func (s Service) WithPrefix(prefix string) blob.Storage {
	return New(s.logger, filepath.Join(prefix, s.prefix), s.wg, s.s3Client, *s.s3Bucket)
}
