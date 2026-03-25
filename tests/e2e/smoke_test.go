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

func doJSON(t *testing.T, method, url string, body interface{}) *http.Response {
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

func parseJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
}

func TestSmoke_FullProjectLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e smoke test")
	}

	ts, cleanup := setupE2E(t)
	defer cleanup()

	// 1. Create a project
	t.Run("create project", func(t *testing.T) {
		resp := doJSON(t, "POST", ts.URL+"/api/projects", map[string]interface{}{
			"name":             "smoke-test-project",
			"description":      "E2E smoke test",
			"repo_url":         "https://github.com/test/smoke",
			"team_composition": []string{"backend-developer"},
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create project: status %d, want 201", resp.StatusCode)
		}

		var p map[string]interface{}
		parseJSON(t, resp, &p)
		if p["name"] != "smoke-test-project" {
			t.Errorf("name = %v", p["name"])
		}
		if p["status"] != "draft" {
			t.Errorf("status = %v, want draft", p["status"])
		}
	})

	// 2. List projects
	t.Run("list projects", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/projects", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list projects: status %d", resp.StatusCode)
		}

		var result map[string]interface{}
		parseJSON(t, resp, &result)
		data := result["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected at least one project")
		}
	})

	// 3. Get project by ID
	var projectID string
	t.Run("get project", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/projects", nil)
		var result map[string]interface{}
		parseJSON(t, resp, &result)
		data := result["data"].([]interface{})
		first := data[0].(map[string]interface{})
		projectID = first["id"].(string)

		resp = doJSON(t, "GET", ts.URL+"/api/projects/"+projectID, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("get project: status %d", resp.StatusCode)
		}
		var p map[string]interface{}
		parseJSON(t, resp, &p)
		if p["id"] != projectID {
			t.Errorf("id mismatch")
		}
	})

	// 4. Update project
	t.Run("update project", func(t *testing.T) {
		resp := doJSON(t, "PATCH", ts.URL+"/api/projects/"+projectID, map[string]string{
			"name":        "smoke-test-updated",
			"description": "Updated description",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("update project: status %d", resp.StatusCode)
		}
		var p map[string]interface{}
		parseJSON(t, resp, &p)
		if p["name"] != "smoke-test-updated" {
			t.Errorf("name = %v, want smoke-test-updated", p["name"])
		}
	})

	// 5. Create a spec
	t.Run("create spec", func(t *testing.T) {
		resp := doJSON(t, "PUT", ts.URL+"/api/projects/"+projectID+"/spec", map[string]interface{}{
			"approved_content": "# Smoke Test Spec\n\nBuild a hello world API.",
			"token_estimate":   50000,
			"agent_count":      1,
		})
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			t.Fatalf("create spec: status %d", resp.StatusCode)
		}
	})

	// 6. Get spec
	t.Run("get spec", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/projects/"+projectID+"/spec", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("get spec: status %d", resp.StatusCode)
		}
		var spec map[string]interface{}
		parseJSON(t, resp, &spec)
		if spec["approval_status"] != "pending" {
			t.Errorf("approval_status = %v", spec["approval_status"])
		}
	})

	// 7. Approve spec
	t.Run("approve spec", func(t *testing.T) {
		resp := doJSON(t, "POST", ts.URL+"/api/projects/"+projectID+"/spec/approve", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("approve spec: status %d", resp.StatusCode)
		}

		// Project status should now be "approved"
		resp = doJSON(t, "GET", ts.URL+"/api/projects/"+projectID, nil)
		var p map[string]interface{}
		parseJSON(t, resp, &p)
		if p["status"] != "approved" {
			t.Errorf("project status = %v, want approved", p["status"])
		}
	})

	// 8. Get risk profile (global default)
	t.Run("get risk profile", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/projects/"+projectID+"/risk-profile", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("get risk profile: status %d", resp.StatusCode)
		}
		var rp map[string]interface{}
		parseJSON(t, resp, &rp)
		if rp["name"] != "Default" {
			t.Errorf("risk profile name = %v, want Default", rp["name"])
		}
	})

	// 9. Get usage (empty)
	t.Run("get usage", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/projects/"+projectID+"/usage", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("get usage: status %d", resp.StatusCode)
		}
	})

	// 10. List agents (empty)
	t.Run("list agents empty", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/projects/"+projectID+"/agents", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list agents: status %d", resp.StatusCode)
		}
	})

	// 11. List tasks (empty)
	t.Run("list tasks empty", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/projects/"+projectID+"/tasks", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list tasks: status %d", resp.StatusCode)
		}
	})

	// 12. Start project (stub — returns 202)
	t.Run("start project stub", func(t *testing.T) {
		resp := doJSON(t, "POST", ts.URL+"/api/projects/"+projectID+"/start", nil)
		if resp.StatusCode != http.StatusAccepted {
			t.Fatalf("start project: status %d, want 202", resp.StatusCode)
		}
	})

	// 13. PO status (stub)
	t.Run("po status stub", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/projects/"+projectID+"/po/status", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("po status: status %d", resp.StatusCode)
		}
	})

	// 14. Health endpoint
	t.Run("health check", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/health", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("health: status %d", resp.StatusCode)
		}
	})

	// 15. Agent library
	t.Run("agent library", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/agents/library", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("library: status %d", resp.StatusCode)
		}
	})

	fmt.Println("Smoke test completed successfully — all endpoints verified")
}
