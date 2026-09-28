package yaml

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

func DecodeFromFilepath(ctx context.Context, filepath string, v any, formatError bool, opts ...yaml.DecodeOption) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	if err = yaml.NewDecoder(file, opts...).DecodeContext(ctx, v); err != nil {
		if formatError {
			return fmt.Errorf("decoding yaml from file: %w", err)
		}
		return fmt.Errorf("decoding yaml from file: %w", errors.New(yaml.FormatError(err, false, false)))
	}

	return nil
}
