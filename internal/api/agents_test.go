package api_test

import (
	"net/http"
	"testing"
)

func TestListAgents_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/bad-id/agents", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestListAgents_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/agents", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestGetAgent_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/bad-id/agents/00000000-0000-0000-0000-000000000002", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetAgent_InvalidAgentID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/agents/not-uuid", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetAgent_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet,
		"/api/projects/00000000-0000-0000-0000-000000000001/agents/00000000-0000-0000-0000-000000000002", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestPauseAgent_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/agents/00000000-0000-0000-0000-000000000002/pause", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPauseAgent_InvalidAgentID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/agents/not-uuid/pause", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPauseAgent_NilBroker_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost,
		"/api/projects/00000000-0000-0000-0000-000000000001/agents/00000000-0000-0000-0000-000000000002/pause", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestResumeAgent_NilBroker_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost,
		"/api/projects/00000000-0000-0000-0000-000000000001/agents/00000000-0000-0000-0000-000000000002/resume", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestStartProject_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/start", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestStartProject_ValidID_Returns202(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/start", "")

	if w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", w.Code, http.StatusAccepted)
	}
}

func TestPauseProject_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/pause", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPauseProject_NilAgentStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/pause", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestResumeProject_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/resume", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestResumeProject_NilAgentStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/resume", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
