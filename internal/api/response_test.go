package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cenron/foundry/internal/api"
	"github.com/cenron/foundry/internal/shared"
)

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	api.RespondJSON(w, http.StatusCreated, map[string]string{"key": "value"})

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("key = %q, want %q", body["key"], "value")
	}
}

func TestRespondError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	api.RespondError(w, &shared.NotFoundError{Resource: "project", ID: "abc"})

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRespondError_Validation(t *testing.T) {
	w := httptest.NewRecorder()
	api.RespondError(w, &shared.ValidationError{Field: "name", Message: "required"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRespondError_Conflict(t *testing.T) {
	w := httptest.NewRecorder()
	api.RespondError(w, &shared.ConflictError{Resource: "spec", Message: "already approved"})

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected error message in response")
	}
}

func TestRespondError_Generic(t *testing.T) {
	w := httptest.NewRecorder()
	api.RespondError(w, errors.New("something went wrong"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
