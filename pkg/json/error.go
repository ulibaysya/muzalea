package json

import (
	"encoding/json"
	"errors"
	"io"
)

// IsDecodingError tells is error related to json decoding process. Error is
// related to decoding process if it matches errors that can be retunred by
// json decoding.
func IsDecodingError(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	} else if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	} else if _, ok := errors.AsType[*json.SyntaxError](err); ok {
		return true
	} else if _, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		return true
	} else if _, ok := errors.AsType[*UnknownFieldErr](err); ok {
		return true
	}
	return false
}
