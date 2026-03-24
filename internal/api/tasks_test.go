package api_test

import (
	"net/http"
	"testing"
)

func TestListTasks_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/bad-id/tasks", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestListTasks_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/tasks", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestListTasks_WithStatusFilter_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/tasks?status=pending", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestGetTask_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/bad-id/tasks/00000000-0000-0000-0000-000000000002", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetTask_InvalidTaskID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/tasks/not-a-uuid", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetTask_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet,
		"/api/projects/00000000-0000-0000-0000-000000000001/tasks/00000000-0000-0000-0000-000000000002", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
