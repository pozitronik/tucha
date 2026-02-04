package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	body := map[string]string{"key": "value"}
	writeJSON(rec, http.StatusCreated, body)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("body key = %q, want %q", got["key"], "value")
	}
}

func TestWriteEnvelope_success(t *testing.T) {
	rec := httptest.NewRecorder()
	writeEnvelope(rec, "user@example.com", 200, "ok")

	if rec.Code != http.StatusOK {
		t.Errorf("HTTP status = %d, want %d", rec.Code, http.StatusOK)
	}

	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Email != "user@example.com" {
		t.Errorf("email = %q", env.Email)
	}
	if env.Status != 200 {
		t.Errorf("status = %d, want 200", env.Status)
	}
	if env.Time <= 0 {
		t.Errorf("time = %d, want positive", env.Time)
	}
	if env.Body != "ok" {
		t.Errorf("body = %v, want %q", env.Body, "ok")
	}
}

func TestWriteEnvelope_errorStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	writeEnvelope(rec, "user@example.com", 400, "bad request")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("HTTP status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Status != 400 {
		t.Errorf("envelope status = %d, want 400", env.Status)
	}
}

func TestWriteSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	writeSuccess(rec, "test@test.com", map[string]int{"count": 42})

	if rec.Code != http.StatusOK {
		t.Errorf("HTTP status = %d, want %d", rec.Code, http.StatusOK)
	}

	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Status != 200 {
		t.Errorf("envelope status = %d, want 200", env.Status)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, "err@test.com", 404, "path", "not found")

	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Body should be {"path": {"error": "not found"}}
	bodyMap, ok := env.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("body is not a map: %T", env.Body)
	}
	pathObj, ok := bodyMap["path"].(map[string]interface{})
	if !ok {
		t.Fatalf("body.path is not a map: %T", bodyMap["path"])
	}
	if msg, ok := pathObj["error"].(string); !ok || msg != "not found" {
		t.Errorf("body.path.error = %v, want %q", pathObj["error"], "not found")
	}
}

func TestWriteHomeError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeHomeError(rec, "user@test.com", 500, "internal error")

	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}

	bodyMap, ok := env.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("body is not a map: %T", env.Body)
	}
	homeObj, ok := bodyMap["home"].(map[string]interface{})
	if !ok {
		t.Fatalf("body.home is not a map: %T", bodyMap["home"])
	}
	if msg, ok := homeObj["error"].(string); !ok || msg != "internal error" {
		t.Errorf("body.home.error = %v, want %q", homeObj["error"], "internal error")
	}
}
