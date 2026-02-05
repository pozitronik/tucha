package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

// Envelope is the standard API v2 response wrapper.
type Envelope struct {
	Email  string      `json:"email"`
	Body   interface{} `json:"body"`
	Time   int64       `json:"time"`
	Status int         `json:"status"`
}

// writeJSON writes a JSON response with the given HTTP status code.
func writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

// writeEnvelope writes a standard API v2 response envelope.
func writeEnvelope(w http.ResponseWriter, email string, status int, body interface{}) {
	env := Envelope{
		Email:  email,
		Body:   body,
		Time:   time.Now().UnixMilli(),
		Status: status,
	}
	httpStatus := http.StatusOK
	if status >= 400 {
		httpStatus = status
	}
	writeJSON(w, httpStatus, env)
}

// writeSuccess writes a successful API v2 response (status 200).
func writeSuccess(w http.ResponseWriter, email string, body interface{}) {
	writeEnvelope(w, email, 200, body)
}

// writeError writes an error API v2 response.
func writeError(w http.ResponseWriter, email string, status int, errField, errMsg string) {
	body := map[string]interface{}{
		errField: map[string]string{
			"error": errMsg,
		},
	}
	writeEnvelope(w, email, status, body)
}

// writeHomeError writes an error response with the error in body.home.error.
func writeHomeError(w http.ResponseWriter, email string, status int, errMsg string) {
	writeError(w, email, status, "home", errMsg)
}

// writeAuthError writes a 403 authentication error response.
func writeAuthError(w http.ResponseWriter) {
	writeEnvelope(w, "", 403, "user")
}
