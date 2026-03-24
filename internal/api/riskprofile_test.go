package api_test

import (
	"net/http"
	"testing"
)

func TestGetRiskProfile_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/bad-id/risk-profile", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetRiskProfile_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/risk-profile", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestUpdateRiskProfile_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPut, "/api/projects/bad-id/risk-profile",
		`{"name":"Custom"}`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateRiskProfile_InvalidJSON_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPut, "/api/projects/00000000-0000-0000-0000-000000000001/risk-profile",
		`not-json`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateRiskProfile_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPut, "/api/projects/00000000-0000-0000-0000-000000000001/risk-profile",
		`{"name":"Custom"}`)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
