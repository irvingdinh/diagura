package http

import (
	"encoding/json"
	nethttp "net/http"
)

// WriteError writes a JSON error response with the given HTTP status code
// and message. The response body is {"error": "<msg>"}.
func WriteError(w nethttp.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
