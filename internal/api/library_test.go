package api_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/api"
)

func TestListLibrary_NilLibrary_Returns200WithEmptyArray(t *testing.T) {
	srv := newTestServer()

	w := doRequest(t, srv, http.MethodGet, "/api/agents/library", "")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body []interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty array, got %d items", len(body))
	}
}

func TestListLibrary_WithLibrary_Returns200WithDefinitions(t *testing.T) {
	dir := t.TempDir()

	// Write a minimal agent definition.
	content := `---
name: backend-developer
description: Builds backend services
tools: all
model: sonnet
---
Backend agent for building APIs.
`
	_ = os.WriteFile(filepath.Join(dir, "backend-developer.md"), []byte(content), 0644)

	lib, err := agent.NewLibrary(dir)
	if err != nil {
		t.Fatalf("NewLibrary() error: %v", err)
	}

	srv := api.NewServer(api.ServerDeps{Library: lib})

	w := doRequest(t, srv, http.MethodGet, "/api/agents/library", "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body []interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body) != 1 {
		t.Errorf("expected 1 definition, got %d", len(body))
	}
}
