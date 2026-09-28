package json

import (
	"encoding/json"
	"fmt"
	"io"
)

func Read[T any](body io.Reader, disallowUnknownFields bool) (T, error) {
	var decoded T
	decoder := json.NewDecoder(body)
	if disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(&decoded); err != nil {
		if fieldErr, ok := AsUnknownFieldErr(err); ok {
			return *new(T), &fieldErr
		}
		return *new(T), err
	}
	return decoded, nil
}

func Write[T any](w io.Writer, v T) error {
	encoded, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding: %w", err)
	}
	if n, err := w.Write(encoded); err != nil {
		return fmt.Errorf("writing: %w", err)
	} else if encodedLen := len(encoded); n != encodedLen {
		return fmt.Errorf("written %v bytes out of %v", n, encodedLen)
	}
	return nil
}
