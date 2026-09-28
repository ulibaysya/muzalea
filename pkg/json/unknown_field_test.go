package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestAsUnknownFieldErr(t *testing.T) {
	const unknownField = "I am an unknown field. I live here. What do you know about me?"

	encodedUnknownField := &strings.Builder{}
	err := json.NewEncoder(encodedUnknownField).Encode(unknownField)
	if err != nil {
		t.Fatalf("encoding unknown field: %v", err)
	}

	unmarshalData := fmt.Sprintf(`{%s:"value"}`, encodedUnknownField)

	decoder := json.NewDecoder(strings.NewReader(unmarshalData))
	decoder.DisallowUnknownFields()

	unmarhsalTo := struct {
		Name string `json:"name"`
	}{}
	if err = decoder.Decode(&unmarhsalTo); err == nil {
		t.Fatalf("expected non-nil error")
	}

	unknownFieldErr, ok := AsUnknownFieldErr(err)
	if !ok {
		t.Fatalf("error not matched unknown field error: %v", err)
	}

	if unknownFieldErr.Field != unknownField {
		t.Fatalf("unexpected unknown field: %v; expected: %s", unknownFieldErr.Field, unknownField)
	}
}

func TestAsUnknownFieldErrRegexMismatch(t *testing.T) {
	err := errors.New("I am completely not unknow field error")

	unknownFieldErr, ok := AsUnknownFieldErr(err)
	if ok {
		t.Fatalf("error unexpectedly matched; expected mismatch: %v", unknownFieldErr)
	}
}

func TestAsUnknownFieldErrImproperlyQuoted(t *testing.T) {
	err := errors.New(`json: unknown field "\"`)

	unknownFieldErr, ok := AsUnknownFieldErr(err)
	if ok {
		t.Fatalf("error unexpectedly matched; expected mismatch: %v", unknownFieldErr)
	}
}
