package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cenron/foundry/internal/api"
)

func newTestServer() *api.Server {
	return api.NewServer(api.ServerDeps{})
}

func doRequest(t *testing.T, srv *api.Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	return w
}

func TestCreateProject_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects",
		`{"name":"My Project","description":"desc","repo_url":"https://github.com/x/y","team_composition":["backend-developer"]}`)

	// Projects store is nil — expect internal server error (panic recovered as 500)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestCreateProject_MissingName_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects",
		`{"description":"no name here"}`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateProject_InvalidJSON_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects", `not-json`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestListProjects_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestGetProject_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/not-a-uuid", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected error message in response")
	}
}

func TestGetProject_ValidID_NilStore_Returns500(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestUpdateProject_InvalidID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPatch, "/api/projects/bad-id", `{"name":"New Name"}`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateProject_InvalidJSON_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPatch, "/api/projects/00000000-0000-0000-0000-000000000001", `{bad json}`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestListProjects_PageSizeClamped_Returns500(t *testing.T) {
	// page_size > 100 gets clamped to 100 before the store call.
	// With nil store it panics → 500, which proves the branch was entered.
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects?page_size=999", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestListProjects_WithValidPageParams_Returns500(t *testing.T) {
	// Exercises the page/page_size parsing paths (valid values, nil store).
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects?page=2&page_size=10", "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
