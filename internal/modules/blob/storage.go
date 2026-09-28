package blob

import (
	"context"
	"io"
)

type Storage interface {
	Save(ctx context.Context, filename string, body io.Reader) error
}
