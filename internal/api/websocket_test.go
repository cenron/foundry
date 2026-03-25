package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleProjectEvents_BadID_Returns400(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/ws/projects/not-a-uuid/events", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	body := strings.TrimSpace(w.Body.String())
	if body == "" {
		t.Error("expected error message in response body")
	}
}

func TestHandleAgentLogs_BadID_Returns400(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/ws/agents/not-a-uuid/logs", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
