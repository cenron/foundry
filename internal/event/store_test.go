package event_test

import (
	"context"
	"os"
	"testing"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/database"
	"github.com/cenron/foundry/internal/event"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/project"
	"github.com/cenron/foundry/internal/shared"
	"github.com/jmoiron/sqlx"
)

func setupTestDB(t *testing.T) *sqlx.DB {
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
		_, _ = db.Exec("DELETE FROM artifacts")
		_, _ = db.Exec("DELETE FROM events")
		_, _ = db.Exec("DELETE FROM tasks")
		_, _ = db.Exec("DELETE FROM agents")
		_, _ = db.Exec("DELETE FROM projects")
		_ = db.Close()
	})

	return db
}

type testFixtures struct {
	project *project.Project
	agent   *agent.Agent
}

func createFixtures(t *testing.T, db *sqlx.DB) testFixtures {
	t.Helper()
	projStore := project.NewStore(db)
	p, err := projStore.Create(context.Background(), project.CreateProjectParams{Name: "Event Test"})
	if err != nil {
		t.Fatalf("creating project: %v", err)
	}

	agentStore := agent.NewStore(db)
	a, err := agentStore.Create(context.Background(), agent.CreateAgentParams{
		ProjectID: p.ID, Role: "backend", Provider: "claude", ContainerID: "c-1",
	})
	if err != nil {
		t.Fatalf("creating agent: %v", err)
	}

	return testFixtures{project: p, agent: a}
}

func TestEventStore_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := event.NewStore(db)
	f := createFixtures(t, db)

	e, err := store.Create(context.Background(), event.CreateEventParams{
		ProjectID: f.project.ID,
		AgentID:   &f.agent.ID,
		Type:      "agent_started",
		Payload:   map[string]string{"role": "backend"},
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if e.Type != "agent_started" {
		t.Errorf("Type = %q, want %q", e.Type, "agent_started")
	}
}

func TestEventStore_ListByProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := event.NewStore(db)
	f := createFixtures(t, db)

	for i := 0; i < 5; i++ {
		_, _ = store.Create(context.Background(), event.CreateEventParams{
			ProjectID: f.project.ID,
			Type:      "task_completed",
			Payload:   map[string]int{"index": i},
		})
	}

	events, total, err := store.ListByProject(context.Background(), f.project.ID, 1, 3)
	if err != nil {
		t.Fatalf("ListByProject() error: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(events) != 3 {
		t.Errorf("len = %d, want 3 (page size)", len(events))
	}
}

func TestEventStore_ListByAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := event.NewStore(db)
	f := createFixtures(t, db)

	agentID := f.agent.ID
	_, _ = store.Create(context.Background(), event.CreateEventParams{
		ProjectID: f.project.ID, AgentID: &agentID, Type: "agent_started", Payload: map[string]string{},
	})
	_, _ = store.Create(context.Background(), event.CreateEventParams{
		ProjectID: f.project.ID, Type: "project_started", Payload: map[string]string{},
	})

	events, err := store.ListByAgent(context.Background(), agentID)
	if err != nil {
		t.Fatalf("ListByAgent() error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("len = %d, want 1", len(events))
	}
}

func TestArtifactStore_CreateAndList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := event.NewArtifactStore(db)
	f := createFixtures(t, db)

	agentID := f.agent.ID
	a, err := store.Create(context.Background(), event.CreateArtifactParams{
		ProjectID:   f.project.ID,
		AgentID:     &agentID,
		Type:        "api_contract",
		Path:        "/shared/contracts/api.yaml",
		Description: "OpenAPI spec for the REST API",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if a.Type != "api_contract" {
		t.Errorf("Type = %q, want %q", a.Type, "api_contract")
	}

	artifacts, err := store.ListByProject(context.Background(), f.project.ID)
	if err != nil {
		t.Fatalf("ListByProject() error: %v", err)
	}
	if len(artifacts) != 1 {
		t.Errorf("len = %d, want 1", len(artifacts))
	}

	// Suppress unused variable
	_ = shared.NewID()
}

func TestArtifactStore_ListByTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	artifactStore := event.NewArtifactStore(db)
	f := createFixtures(t, db)

	// Create a task to associate artifacts with.
	taskStore := orchestrator.NewTaskStore(db)
	task, err := taskStore.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: f.project.ID,
		Title:     "Artifact task",
	})
	if err != nil {
		t.Fatalf("creating task: %v", err)
	}

	agentID := f.agent.ID

	// Artifact belonging to the task.
	_, err = artifactStore.Create(context.Background(), event.CreateArtifactParams{
		ProjectID:   f.project.ID,
		TaskID:      &task.ID,
		AgentID:     &agentID,
		Type:        "code",
		Path:        "/src/main.go",
		Description: "Main entrypoint",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Artifact with no task (should not appear in the task-scoped query).
	_, _ = artifactStore.Create(context.Background(), event.CreateArtifactParams{
		ProjectID:   f.project.ID,
		AgentID:     &agentID,
		Type:        "docs",
		Path:        "/docs/readme.md",
		Description: "README",
	})

	artifacts, err := artifactStore.ListByTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("ListByTask() error: %v", err)
	}

	if len(artifacts) != 1 {
		t.Errorf("len = %d, want 1", len(artifacts))
	}
	if len(artifacts) > 0 && artifacts[0].Path != "/src/main.go" {
		t.Errorf("Path = %q, want %q", artifacts[0].Path, "/src/main.go")
	}
}
