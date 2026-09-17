package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// DecodeJSON decodes r.Body into dst. Oversized bodies (MaxBytesReader) return 413.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dst)
	if err == nil {
		return true
	}
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		WriteBodyTooLarge(w)
		return false
	}
	// MaxBytesReader may surface as io error wrapping
	if errors.Is(err, io.ErrUnexpectedEOF) {
		// still treat generic decode errors as bad request unless MaxBytes
	}
	Fail(w, 400, CodeBadRequest, "invalid JSON body")
	return false
}
