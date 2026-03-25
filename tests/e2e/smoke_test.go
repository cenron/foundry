package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/api"
	"github.com/cenron/foundry/internal/database"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/project"
)

func testDatabaseURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://foundry:foundry@localhost:5433/foundry?sslmode=disable"
	}
	return url
}

func setupE2E(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	db, err := database.Connect(context.Background(), testDatabaseURL())
	if err != nil {
		t.Skipf("skipping e2e: database not available: %v", err)
	}

	if err := database.MigrateUp(db, "../../migrations"); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	srv := api.NewServer(api.ServerDeps{
		Projects:     project.NewStore(db),
		Specs:        project.NewSpecStore(db),
		Tasks:        orchestrator.NewTaskStore(db),
		Agents:       agent.NewStore(db),
		RiskProfiles: project.NewRiskProfileStore(db),
	})

	ts := httptest.NewServer(srv.Handler())

	cleanup := func() {
		ts.Close()
		_, _ = db.Exec("DELETE FROM events")
		_, _ = db.Exec("DELETE FROM artifacts")
		_, _ = db.Exec("DELETE FROM agent_messages")
		_, _ = db.Exec("DELETE FROM tasks")
		_, _ = db.Exec("DELETE FROM agents")
		_, _ = db.Exec("DELETE FROM specs")
		_, _ = db.Exec("DELETE FROM po_sessions")
		_, _ = db.Exec("DELETE FROM projects")
		_ = db.Close()
	}

	return ts, cleanup
}

// doRequest executes an HTTP request and returns the status code.
// The response body is always closed.
func doRequest(t *testing.T, method, url string, body interface{}) int {
	t.Helper()
	resp := execRequest(t, method, url, body)
	_ = resp.Body.Close()
	return resp.StatusCode
}

// doRequestJSON executes an HTTP request, decodes the response body into v,
// and returns the status code. The response body is always closed.
func doRequestJSON(t *testing.T, method, url string, body interface{}, v interface{}) int {
	t.Helper()
	resp := execRequest(t, method, url, body)
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	return resp.StatusCode
}

// execRequest builds and sends an HTTP request, returning the raw response.
// Callers are responsible for closing resp.Body.
func execRequest(t *testing.T, method, url string, body interface{}) *http.Response {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("sending request: %v", err)
	}
	return resp
}

func TestSmoke_FullProjectLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e smoke test")
	}

	ts, cleanup := setupE2E(t)
	defer cleanup()

	// 1. Create a project
	t.Run("create project", func(t *testing.T) {
		var p map[string]interface{}
		status := doRequestJSON(t, "POST", ts.URL+"/api/projects", map[string]interface{}{
			"name":             "smoke-test-project",
			"description":      "E2E smoke test",
			"repo_url":         "https://github.com/test/smoke",
			"team_composition": []string{"backend-developer"},
		}, &p)
		if status != http.StatusCreated {
			t.Fatalf("create project: status %d, want 201", status)
		}

		if p["name"] != "smoke-test-project" {
			t.Errorf("name = %v", p["name"])
		}
		if p["status"] != "draft" {
			t.Errorf("status = %v, want draft", p["status"])
		}
	})

	// 2. List projects
	t.Run("list projects", func(t *testing.T) {
		var result map[string]interface{}
		status := doRequestJSON(t, "GET", ts.URL+"/api/projects", nil, &result)
		if status != http.StatusOK {
			t.Fatalf("list projects: status %d", status)
		}

		data := result["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected at least one project")
		}
	})

	// 3. Get project by ID
	var projectID string
	t.Run("get project", func(t *testing.T) {
		var result map[string]interface{}
		doRequestJSON(t, "GET", ts.URL+"/api/projects", nil, &result)
		data := result["data"].([]interface{})
		first := data[0].(map[string]interface{})
		projectID = first["id"].(string)

		var p map[string]interface{}
		status := doRequestJSON(t, "GET", ts.URL+"/api/projects/"+projectID, nil, &p)
		if status != http.StatusOK {
			t.Fatalf("get project: status %d", status)
		}
		if p["id"] != projectID {
			t.Errorf("id mismatch")
		}
	})

	// 4. Update project
	t.Run("update project", func(t *testing.T) {
		var p map[string]interface{}
		status := doRequestJSON(t, "PATCH", ts.URL+"/api/projects/"+projectID, map[string]string{
			"name":        "smoke-test-updated",
			"description": "Updated description",
		}, &p)
		if status != http.StatusOK {
			t.Fatalf("update project: status %d", status)
		}
		if p["name"] != "smoke-test-updated" {
			t.Errorf("name = %v, want smoke-test-updated", p["name"])
		}
	})

	// 5. Create a spec
	t.Run("create spec", func(t *testing.T) {
		status := doRequest(t, "PUT", ts.URL+"/api/projects/"+projectID+"/spec", map[string]interface{}{
			"approved_content": "# Smoke Test Spec\n\nBuild a hello world API.",
			"token_estimate":   50000,
			"agent_count":      1,
		})
		if status != http.StatusOK && status != http.StatusCreated {
			t.Fatalf("create spec: status %d", status)
		}
	})

	// 6. Get spec
	t.Run("get spec", func(t *testing.T) {
		var spec map[string]interface{}
		status := doRequestJSON(t, "GET", ts.URL+"/api/projects/"+projectID+"/spec", nil, &spec)
		if status != http.StatusOK {
			t.Fatalf("get spec: status %d", status)
		}
		if spec["approval_status"] != "pending" {
			t.Errorf("approval_status = %v", spec["approval_status"])
		}
	})

	// 7. Approve spec
	t.Run("approve spec", func(t *testing.T) {
		status := doRequest(t, "POST", ts.URL+"/api/projects/"+projectID+"/spec/approve", nil)
		if status != http.StatusOK {
			t.Fatalf("approve spec: status %d", status)
		}

		// Project status should now be "approved"
		var p map[string]interface{}
		doRequestJSON(t, "GET", ts.URL+"/api/projects/"+projectID, nil, &p)
		if p["status"] != "approved" {
			t.Errorf("project status = %v, want approved", p["status"])
		}
	})

	// 8. Get risk profile (global default)
	t.Run("get risk profile", func(t *testing.T) {
		var rp map[string]interface{}
		status := doRequestJSON(t, "GET", ts.URL+"/api/projects/"+projectID+"/risk-profile", nil, &rp)
		if status != http.StatusOK {
			t.Fatalf("get risk profile: status %d", status)
		}
		if rp["name"] != "Default" {
			t.Errorf("risk profile name = %v, want Default", rp["name"])
		}
	})

	// 9. Get usage (empty)
	t.Run("get usage", func(t *testing.T) {
		status := doRequest(t, "GET", ts.URL+"/api/projects/"+projectID+"/usage", nil)
		if status != http.StatusOK {
			t.Fatalf("get usage: status %d", status)
		}
	})

	// 10. List agents (empty)
	t.Run("list agents empty", func(t *testing.T) {
		status := doRequest(t, "GET", ts.URL+"/api/projects/"+projectID+"/agents", nil)
		if status != http.StatusOK {
			t.Fatalf("list agents: status %d", status)
		}
	})

	// 11. List tasks (empty)
	t.Run("list tasks empty", func(t *testing.T) {
		status := doRequest(t, "GET", ts.URL+"/api/projects/"+projectID+"/tasks", nil)
		if status != http.StatusOK {
			t.Fatalf("list tasks: status %d", status)
		}
	})

	// 12. Start project (stub — returns 202)
	t.Run("start project stub", func(t *testing.T) {
		status := doRequest(t, "POST", ts.URL+"/api/projects/"+projectID+"/start", nil)
		if status != http.StatusAccepted {
			t.Fatalf("start project: status %d, want 202", status)
		}
	})

	// 13. PO status (stub)
	t.Run("po status stub", func(t *testing.T) {
		status := doRequest(t, "GET", ts.URL+"/api/projects/"+projectID+"/po/status", nil)
		if status != http.StatusOK {
			t.Fatalf("po status: status %d", status)
		}
	})

	// 14. Health endpoint
	t.Run("health check", func(t *testing.T) {
		status := doRequest(t, "GET", ts.URL+"/api/health", nil)
		if status != http.StatusOK {
			t.Fatalf("health: status %d", status)
		}
	})

	// 15. Agent library
	t.Run("agent library", func(t *testing.T) {
		status := doRequest(t, "GET", ts.URL+"/api/agents/library", nil)
		if status != http.StatusOK {
			t.Fatalf("library: status %d", status)
		}
	})

	fmt.Println("Smoke test completed successfully — all endpoints verified")
}
