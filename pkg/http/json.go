package http

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
)

func ErrorJSON(w http.ResponseWriter, error string, code int) {
	h := w.Header()

	h.Del("Content-Length")

	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.MarshalWrite(w, error) //nolint:errcheck
}

func FmtErrorJSON(w http.ResponseWriter, code int, format string, a ...any) {
	ErrorJSON(w, fmt.Sprintf(format, a...), code)
}

func ErrorJSON400(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func ErrorJSON401(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func ErrorJSON404(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func ErrorJSON409(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusText(http.StatusConflict), http.StatusConflict)
}

func ErrorJSON413(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
}

func ErrorJSON500(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
