package json

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	// TODO there can't be newlines in encoded go string, so we shouldn't add
	// this flag, correct?
	// regexUnknownField = regexp.MustCompile(`^json: unknown field "(?s).*"$`)
	regexUnknownField = regexp.MustCompile(`^json: unknown field ".*"$`)
)

type UnknownFieldErr struct {
	Field string
}

func (e UnknownFieldErr) Error() string {
	return fmt.Sprintf("json: unknown field %q", e.Field)
}

// AsUnknownFieldErr accepts an error and checks if that error matches a format
// of an error caused by [json.Decoder.DisallowUnknownFields] and
// [json.Decoder.Decode]. If so, AsUnknownFieldErr returns [UnknownFieldErr]
// with unknown field set and true. Otherwise zero value for [UnknownFieldErr]
// and false are returned.
func AsUnknownFieldErr(err error) (UnknownFieldErr, bool) {
	errStr := err.Error()

	if !regexUnknownField.MatchString(errStr) {
		return UnknownFieldErr{}, false
	}

	const firstQuoteIndex = 20

	actualField, err := strconv.Unquote(errStr[firstQuoteIndex:])
	if err != nil {
		return UnknownFieldErr{}, false
	}

	return UnknownFieldErr{
		Field: actualField,
	}, true
}
