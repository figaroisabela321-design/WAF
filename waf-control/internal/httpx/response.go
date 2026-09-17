package httpx

import (
	"encoding/json"
	"net/http"
)

// Envelope is the unified JSON response format.
type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// WriteJSON writes a JSON response with the given status and envelope.
func WriteJSON(w http.ResponseWriter, status int, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// OK writes a success envelope (code 0).
func OK(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, 0, "ok", data)
}

// Fail writes an error envelope with HTTP status and business code.
func Fail(w http.ResponseWriter, status int, code int, message string) {
	WriteJSON(w, status, code, message, nil)
}

// EncodeEnvelope encodes an Envelope to JSON bytes (for tests).
func EncodeEnvelope(code int, message string, data any) ([]byte, error) {
	return json.Marshal(Envelope{Code: code, Message: message, Data: data})
}
