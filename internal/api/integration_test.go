package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/api"
	"github.com/cenron/foundry/internal/database"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/project"
	"github.com/jmoiron/sqlx"
)

func setupIntegrationDB(t *testing.T) *sqlx.DB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://foundry:foundry@localhost:5433/foundry?sslmode=disable"
	}

	db, err := database.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connecting to test db: %v", err)
	}

	if err := database.MigrateUp(db, "../../migrations"); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM risk_profiles WHERE project_id IS NOT NULL")
		_, _ = db.Exec("DELETE FROM artifacts")
		_, _ = db.Exec("DELETE FROM events")
		_, _ = db.Exec("DELETE FROM tasks")
		_, _ = db.Exec("DELETE FROM agents")
		_, _ = db.Exec("DELETE FROM specs")
		_, _ = db.Exec("DELETE FROM projects")
		_ = db.Close()
	})

	return db
}

func setupIntegrationServer(t *testing.T) (*api.Server, *sqlx.DB) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupIntegrationDB(t)
	srv := api.NewServer(api.ServerDeps{
		Projects:     project.NewStore(db),
		Specs:        project.NewSpecStore(db),
		Tasks:        orchestrator.NewTaskStore(db),
		Agents:       agent.NewStore(db),
		RiskProfiles: project.NewRiskProfileStore(db),
	})
	return srv, db
}

// --- Projects ---

func TestIntegration_CreateProject_Returns201(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	w := doRequest(t, srv, http.MethodPost, "/api/projects",
		`{"name":"Integration Project","description":"desc","repo_url":"https://github.com/x/y","team_composition":["backend-developer"]}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["id"] == nil {
		t.Error("expected id in response")
	}
	if body["name"] != "Integration Project" {
		t.Errorf("name = %v, want %q", body["name"], "Integration Project")
	}
}

func TestIntegration_ListProjects_Returns200(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	// Create a project first so list is non-empty.
	doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"List Me"}`)

	w := doRequest(t, srv, http.MethodGet, "/api/projects", "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["data"] == nil {
		t.Error("expected data field in response")
	}
}

func TestIntegration_GetProject_Returns200(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	// Create the project.
	createResp := doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"Get Me"}`)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("setup: create project failed with %d", createResp.Code)
	}

	var created map[string]interface{}
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}
	id := created["id"].(string)

	w := doRequest(t, srv, http.MethodGet, fmt.Sprintf("/api/projects/%s", id), "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["id"] != id {
		t.Errorf("id = %v, want %q", body["id"], id)
	}
}

func TestIntegration_GetProject_NotFound_Returns404(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	w := doRequest(t, srv, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000099", "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestIntegration_UpdateProject_Returns200(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	createResp := doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"Before Update"}`)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("setup: create failed with %d", createResp.Code)
	}

	var created map[string]interface{}
	_ = json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	w := doRequest(t, srv, http.MethodPatch, fmt.Sprintf("/api/projects/%s", id),
		`{"name":"After Update","description":"new desc"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["name"] != "After Update" {
		t.Errorf("name = %v, want %q", body["name"], "After Update")
	}
}

func TestIntegration_UpdateProject_NotFound_Returns404(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	w := doRequest(t, srv, http.MethodPatch, "/api/projects/00000000-0000-0000-0000-000000000099",
		`{"name":"Ghost"}`)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- Specs ---

func TestIntegration_UpdateAndGetSpec_Returns200(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	createResp := doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"Spec Project"}`)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("setup: create project failed with %d", createResp.Code)
	}

	var created map[string]interface{}
	_ = json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	putW := doRequest(t, srv, http.MethodPut, fmt.Sprintf("/api/projects/%s/spec", id),
		`{"approved_content":"# Spec","token_estimate":50000,"agent_count":3}`)

	if putW.Code != http.StatusOK {
		t.Fatalf("update spec status = %d, want %d; body: %s", putW.Code, http.StatusOK, putW.Body.String())
	}

	getW := doRequest(t, srv, http.MethodGet, fmt.Sprintf("/api/projects/%s/spec", id), "")

	if getW.Code != http.StatusOK {
		t.Fatalf("get spec status = %d, want %d; body: %s", getW.Code, http.StatusOK, getW.Body.String())
	}

	var spec map[string]interface{}
	if err := json.NewDecoder(getW.Body).Decode(&spec); err != nil {
		t.Fatalf("decoding spec: %v", err)
	}
	if spec["approved_content"] != "# Spec" {
		t.Errorf("approved_content = %v, want %q", spec["approved_content"], "# Spec")
	}
}

func TestIntegration_ApproveSpec_Returns200(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	// Create project + spec.
	createResp := doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"Approve Project"}`)
	var created map[string]interface{}
	_ = json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	doRequest(t, srv, http.MethodPut, fmt.Sprintf("/api/projects/%s/spec", id),
		`{"approved_content":"# Spec","token_estimate":50000,"agent_count":3}`)

	w := doRequest(t, srv, http.MethodPost, fmt.Sprintf("/api/projects/%s/spec/approve", id), "")

	if w.Code != http.StatusOK {
		t.Fatalf("approve spec status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var spec map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&spec); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if spec["approval_status"] != "approved" {
		t.Errorf("approval_status = %v, want %q", spec["approval_status"], "approved")
	}
}

func TestIntegration_ApproveSpec_MissingContent_Returns400(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	createResp := doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"Empty Spec Project"}`)
	var created map[string]interface{}
	_ = json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	// Spec with no content.
	doRequest(t, srv, http.MethodPut, fmt.Sprintf("/api/projects/%s/spec", id),
		`{"approved_content":"","token_estimate":0}`)

	w := doRequest(t, srv, http.MethodPost, fmt.Sprintf("/api/projects/%s/spec/approve", id), "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestIntegration_ApproveSpec_MissingTokenEstimate_Returns400(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	createResp := doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"No Estimate Project"}`)
	var created map[string]interface{}
	_ = json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	// Spec with content but zero token estimate.
	doRequest(t, srv, http.MethodPut, fmt.Sprintf("/api/projects/%s/spec", id),
		`{"approved_content":"# My Spec","token_estimate":0}`)

	w := doRequest(t, srv, http.MethodPost, fmt.Sprintf("/api/projects/%s/spec/approve", id), "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

// --- Tasks ---

func TestIntegration_ListTasks_Returns200(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Task Project"})

	taskStore := orchestrator.NewTaskStore(db)
	_, _ = taskStore.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: p.ID, Title: "Task 1",
	})
	_, _ = taskStore.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: p.ID, Title: "Task 2",
	})

	w := doRequest(t, srv, http.MethodGet, fmt.Sprintf("/api/projects/%s/tasks", p.ID), "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var tasks []interface{}
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("len = %d, want 2", len(tasks))
	}
}

func TestIntegration_ListTasks_WithStatusFilter_Returns200(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Filter Task Project"})

	taskStore := orchestrator.NewTaskStore(db)
	task1, _ := taskStore.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: p.ID, Title: "Pending Task",
	})
	_, _ = taskStore.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: p.ID, Title: "Other Task",
	})
	_ = taskStore.UpdateStatus(context.Background(), task1.ID, "done")

	w := doRequest(t, srv, http.MethodGet,
		fmt.Sprintf("/api/projects/%s/tasks?status=done", p.ID), "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var tasks []interface{}
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("len = %d, want 1 (only 'done' tasks)", len(tasks))
	}
}

// --- Agents ---

func TestIntegration_ListAgents_Returns200(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Agent Project"})

	agentStore := agent.NewStore(db)
	_, _ = agentStore.Create(context.Background(), agent.CreateAgentParams{
		ProjectID: p.ID, Role: "backend", Provider: "claude", ContainerID: "c-1",
	})

	w := doRequest(t, srv, http.MethodGet, fmt.Sprintf("/api/projects/%s/agents", p.ID), "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var agents []interface{}
	if err := json.NewDecoder(w.Body).Decode(&agents); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("len = %d, want 1", len(agents))
	}
}

// --- Usage ---

func TestIntegration_GetUsage_Returns200(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Usage Project"})

	w := doRequest(t, srv, http.MethodGet, fmt.Sprintf("/api/projects/%s/usage", p.ID), "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["total_tokens"] == nil {
		t.Error("expected total_tokens in response")
	}
}

// --- Risk Profile ---

func TestIntegration_GetRiskProfile_Returns200(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Risk Profile Project"})

	w := doRequest(t, srv, http.MethodGet, fmt.Sprintf("/api/projects/%s/risk-profile", p.ID), "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["name"] == nil {
		t.Error("expected name in risk profile response")
	}
}

func TestIntegration_UpdateRiskProfile_CreatesProjectSpecific(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Risk Update Project"})

	// Updating a project that has no project-specific profile should create one.
	w := doRequest(t, srv, http.MethodPut, fmt.Sprintf("/api/projects/%s/risk-profile", p.ID),
		`{"name":"Custom Risk","low_criteria":{},"medium_criteria":{},"high_criteria":{},"model_routing":{}}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["name"] != "Custom Risk" {
		t.Errorf("name = %v, want %q", body["name"], "Custom Risk")
	}
	// project_id should be set on the newly created profile.
	if body["project_id"] == nil {
		t.Error("expected project_id to be set on created profile")
	}
}

func TestIntegration_UpdateRiskProfile_UpdatesExistingProjectSpecific(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Risk Update Existing"})

	// First PUT creates a project-specific profile.
	doRequest(t, srv, http.MethodPut, fmt.Sprintf("/api/projects/%s/risk-profile", p.ID),
		`{"name":"First Name","low_criteria":{},"medium_criteria":{},"high_criteria":{},"model_routing":{}}`)

	// Second PUT should update the existing profile.
	w := doRequest(t, srv, http.MethodPut, fmt.Sprintf("/api/projects/%s/risk-profile", p.ID),
		`{"name":"Second Name","low_criteria":{},"medium_criteria":{},"high_criteria":{},"model_routing":{}}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["name"] != "Second Name" {
		t.Errorf("name = %v, want %q", body["name"], "Second Name")
	}
}

// --- Library (no DB needed) ---

func TestIntegration_Library_Returns200WithEmptyArray(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	w := doRequest(t, srv, http.MethodGet, "/api/agents/library", "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// --- Reject Spec ---

func TestIntegration_RejectSpec_Returns200(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	createResp := doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"Reject Project"}`)
	var created map[string]interface{}
	_ = json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	doRequest(t, srv, http.MethodPut, fmt.Sprintf("/api/projects/%s/spec", id),
		`{"approved_content":"# Spec","token_estimate":50000,"agent_count":3}`)

	w := doRequest(t, srv, http.MethodPost, fmt.Sprintf("/api/projects/%s/spec/reject", id), "")

	if w.Code != http.StatusOK {
		t.Fatalf("reject spec status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var spec map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&spec); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if spec["approval_status"] != "rejected" {
		t.Errorf("approval_status = %v, want %q", spec["approval_status"], "rejected")
	}
}

// --- Spec — GetSpec not found ---

func TestIntegration_GetSpec_NotFound_Returns404(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	// Create project but no spec.
	createResp := doRequest(t, srv, http.MethodPost, "/api/projects", `{"name":"No Spec Project"}`)
	var created map[string]interface{}
	_ = json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	w := doRequest(t, srv, http.MethodGet, fmt.Sprintf("/api/projects/%s/spec", id), "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

// --- publishProjectCommand path (pause/resume with real agent store) ---

func TestIntegration_PauseProject_NoBroker_Returns500(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Pause Project"})

	agentStore := agent.NewStore(db)
	_, _ = agentStore.Create(context.Background(), agent.CreateAgentParams{
		ProjectID: p.ID, Role: "backend", Provider: "claude", ContainerID: "c-1",
	})

	// srv has no broker configured → publishProjectCommand returns 500 after listing agents.
	w := doRequest(t, srv, http.MethodPost, fmt.Sprintf("/api/projects/%s/pause", p.ID), "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (no broker); body: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_PauseProject_NoAgents_Returns500(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Empty Pause Project"})

	// No agents — publishProjectCommand with no broker should still return 500 (broker == nil check).
	w := doRequest(t, srv, http.MethodPost, fmt.Sprintf("/api/projects/%s/pause", p.ID), "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body: %s", w.Code, w.Body.String())
	}
}

// --- GetAgent with a real store (404 path) ---

func TestIntegration_GetAgent_NotFound_Returns404(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	w := doRequest(t, srv, http.MethodGet,
		"/api/projects/00000000-0000-0000-0000-000000000001/agents/00000000-0000-0000-0000-000000000099", "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

// --- GetTask 404 ---

func TestIntegration_GetTask_NotFound_Returns404(t *testing.T) {
	srv, _ := setupIntegrationServer(t)

	w := doRequest(t, srv, http.MethodGet,
		"/api/projects/00000000-0000-0000-0000-000000000001/tasks/00000000-0000-0000-0000-000000000099", "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

// --- GetUsage with tasks that have token usage ---

func TestIntegration_GetUsage_WithTasks_ReturnsTokenBreakdown(t *testing.T) {
	srv, db := setupIntegrationServer(t)

	projStore := project.NewStore(db)
	p, _ := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Token Usage Project"})

	taskStore := orchestrator.NewTaskStore(db)
	task, _ := taskStore.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: p.ID, Title: "Tokenized Task",
	})
	// Manually set token usage via UpdateStatus (just to ensure the task exists)
	_ = taskStore.UpdateStatus(context.Background(), task.ID, "done")

	w := doRequest(t, srv, http.MethodGet, fmt.Sprintf("/api/projects/%s/usage", p.ID), "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	breakdown, ok := body["task_breakdown"].([]interface{})
	if !ok {
		t.Fatal("expected task_breakdown array in response")
	}
	if len(breakdown) != 1 {
		t.Errorf("task_breakdown len = %d, want 1", len(breakdown))
	}
}
