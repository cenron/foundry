package api_test

import (
	"net/http"
	"testing"
)

func TestGetSpec_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/bad-id/spec", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetSpec_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/spec", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestUpdateSpec_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPut, "/api/projects/bad-id/spec",
		`{"approved_content":"# Spec","token_estimate":50000,"agent_count":3}`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateSpec_InvalidJSON_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPut, "/api/projects/00000000-0000-0000-0000-000000000001/spec", `not-json`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateSpec_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPut, "/api/projects/00000000-0000-0000-0000-000000000001/spec",
		`{"approved_content":"# Spec","token_estimate":50000,"agent_count":3}`)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestApproveSpec_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/spec/approve", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestApproveSpec_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/spec/approve", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestRejectSpec_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/spec/reject", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRejectSpec_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/spec/reject", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
