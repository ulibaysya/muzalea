// source: https://victoronsoftware.com/posts/go-json-unmarshal/

package json

import (
	"bytes"
	"encoding/json"
)

type Any[T any] struct {
	Set   bool
	Null  bool
	Value T
}

// MarshalJSON implements the json.Marshaler interface.
// Only Value is marshaled, and only if Valid is true.
func (s Any[T]) MarshalJSON() ([]byte, error) {
	if !s.Null {
		return []byte("null"), nil
	}
	return json.Marshal(s.Value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Set is always set to true, even if the JSON data was set to null.
// Valid is set if the JSON data is not set to null.
func (s *Any[T]) UnmarshalJSON(data []byte) error {
	s.Set = true
	s.Null = false

	if bytes.Equal(data, []byte("null")) {
		var zero T
		s.Value = zero
		return nil
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	s.Value = v
	s.Null = true
	return nil
}
