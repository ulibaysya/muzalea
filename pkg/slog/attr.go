package slog

import (
	"log/slog"
)

func Error(key string, value error) slog.Attr {
	if wrappedErr, ok := value.(interface{ Unwrap() []error }); ok {
		errs := wrappedErr.Unwrap()
		strErrors := make([]string, len(errs))
		for i, err := range errs {
			strErrors[i] = err.Error()
		}
		return slog.Any(key, strErrors)
	}
	return slog.Any(key, value)
}
