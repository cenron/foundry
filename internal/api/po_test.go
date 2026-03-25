package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cenron/foundry/internal/api"
	"github.com/cenron/foundry/internal/po"
)

// newPOServer builds a test server with a SessionManager injected so the
// PO != nil code paths are exercised.
func newPOServer(t *testing.T) *api.Server {
	t.Helper()
	sm := po.NewSessionManager(t.TempDir(), "test-key", "latest")
	return api.NewServer(api.ServerDeps{PO: sm})
}

func TestPOChat_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/po/chat", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPOChat_ValidID_Returns501(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/po/chat", "")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected error message in response")
	}
}

func TestPOChatDelete_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodDelete, "/api/projects/bad-id/po/chat", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPOChatDelete_ValidID_Returns501(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodDelete, "/api/projects/00000000-0000-0000-0000-000000000001/po/chat", "")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}

func TestPOStatus_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/bad-id/po/status", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPOStatus_ValidID_Returns200WithInactive(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/po/status", "")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["active"] != false {
		t.Errorf("active = %v, want false", body["active"])
	}
}

func TestPOPlanning_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/po/planning", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPOPlanning_ValidID_Returns501(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/po/planning", "")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}

func TestPOEstimation_InvalidProjectID_Returns400(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/bad-id/po/estimation", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPOEstimation_ValidID_Returns501(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/po/estimation", "")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}

// --- PO configured (PO != nil) paths ---

func TestPOStatus_WithPO_ReturnsActiveTrue(t *testing.T) {
	// Inject a fake session so IsActive returns true.
	sm := po.NewSessionManager(t.TempDir(), "key", "latest")
	sm.InjectSession("00000000-0000-0000-0000-000000000001", &po.POSession{
		ID:          "test-session",
		ProjectName: "00000000-0000-0000-0000-000000000001",
		Status:      po.SessionStatusActive,
	})
	srv := api.NewServer(api.ServerDeps{PO: sm})

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/po/status", "")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["active"] != true {
		t.Errorf("active = %v, want true", body["active"])
	}
}

func TestPOStatus_WithPO_ReturnsActiveFalse(t *testing.T) {
	srv := newPOServer(t)

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000001/po/status", "")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["active"] != false {
		t.Errorf("active = %v, want false", body["active"])
	}
}

func TestPOChat_WithPO_InvalidJSON_Returns400(t *testing.T) {
	srv := newPOServer(t)

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/po/chat", "not-json")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPOChat_WithPO_ValidMessage_StartSessionFails(t *testing.T) {
	// StartSession will fail because no 'claude' binary exists in CI/test env.
	// The handler must return a non-2xx status (not 501) — verifies the PO != nil path.
	srv := newPOServer(t)

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/po/chat",
		`{"message":"hello"}`)

	// Should not be 501 (that's the nil-PO path) — any other error code is acceptable.
	if w.Code == http.StatusNotImplemented {
		t.Errorf("status = 501, but PO is configured — should not return 501")
	}
}

func TestPOPlanning_WithPO_StartSessionFails(t *testing.T) {
	srv := newPOServer(t)

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/po/planning", "")

	if w.Code == http.StatusNotImplemented {
		t.Errorf("status = 501, but PO is configured — should not return 501")
	}
}

func TestPOEstimation_WithPO_StartSessionFails(t *testing.T) {
	srv := newPOServer(t)

	w := doRequest(t, srv, http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000001/po/estimation", "")

	if w.Code == http.StatusNotImplemented {
		t.Errorf("status = 501, but PO is configured — should not return 501")
	}
}

func TestPOChatDelete_WithPO_NoActiveSession_ReturnsError(t *testing.T) {
	srv := newPOServer(t)

	// No session injected — StopSession returns an error.
	w := doRequest(t, srv, http.MethodDelete, "/api/projects/00000000-0000-0000-0000-000000000001/po/chat", "")

	// Should not be 501 — PO is configured.
	if w.Code == http.StatusNotImplemented {
		t.Errorf("status = 501, but PO is configured — should not return 501")
	}
}

func TestPOChatDelete_WithPO_ActiveSession_Returns200(t *testing.T) {
	sm := po.NewSessionManager(t.TempDir(), "key", "latest")
	sm.InjectSession("00000000-0000-0000-0000-000000000001", &po.POSession{
		ID:          "session-id",
		ProjectName: "00000000-0000-0000-0000-000000000001",
		Status:      po.SessionStatusActive,
	})
	srv := api.NewServer(api.ServerDeps{PO: sm})

	w := doRequest(t, srv, http.MethodDelete, "/api/projects/00000000-0000-0000-0000-000000000001/po/chat", "")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["status"] != "session stopped" {
		t.Errorf("status = %q, want %q", body["status"], "session stopped")
	}
}
