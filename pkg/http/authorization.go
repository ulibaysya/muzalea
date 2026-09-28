package http

import (
	"net/http"
	"strings"
)

func ExtractAuthorization(h http.Header) (string, string, bool) {
	authHeaders := h.Values("Authorization")
	if len(authHeaders) != 1 {
		return "", "", false
	}

	headerFields := strings.Fields(authHeaders[0])
	if len(headerFields) != 2 {
		return "", "", false
	}

	return headerFields[0], headerFields[1], true
}
