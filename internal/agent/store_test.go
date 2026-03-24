package agent_test

import (
	"context"
	"os"
	"testing"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/database"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/project"
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
		_, _ = db.Exec("DELETE FROM tasks")
		_, _ = db.Exec("DELETE FROM agents")
		_, _ = db.Exec("DELETE FROM projects")
		_ = db.Close()
	})

	return db
}

func createTestProject(t *testing.T, db *sqlx.DB) *project.Project {
	t.Helper()
	store := project.NewStore(db)
	p, err := store.Create(context.Background(), project.CreateProjectParams{Name: "Agent Test"})
	if err != nil {
		t.Fatalf("creating test project: %v", err)
	}
	return p
}

func TestAgentStore_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := agent.NewStore(db)
	proj := createTestProject(t, db)

	a, err := store.Create(context.Background(), agent.CreateAgentParams{
		ProjectID:   proj.ID,
		Role:        "backend-developer",
		Provider:    "claude",
		ContainerID: "container-123",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if a.Role != "backend-developer" {
		t.Errorf("Role = %q, want %q", a.Role, "backend-developer")
	}
	if a.Status != "starting" {
		t.Errorf("Status = %q, want %q", a.Status, "starting")
	}
	if a.Health != "healthy" {
		t.Errorf("Health = %q, want %q", a.Health, "healthy")
	}
}

func TestAgentStore_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := agent.NewStore(db)
	proj := createTestProject(t, db)

	created, _ := store.Create(context.Background(), agent.CreateAgentParams{
		ProjectID: proj.ID, Role: "qa", Provider: "claude", ContainerID: "c-1",
	})

	got, err := store.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID mismatch")
	}
}

func TestAgentStore_ListByProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := agent.NewStore(db)
	proj := createTestProject(t, db)

	roles := []string{"backend-developer", "frontend-developer", "qa"}
	for _, role := range roles {
		_, _ = store.Create(context.Background(), agent.CreateAgentParams{
			ProjectID: proj.ID, Role: role, Provider: "claude", ContainerID: "c-1",
		})
	}

	agents, err := store.ListByProject(context.Background(), proj.ID)
	if err != nil {
		t.Fatalf("ListByProject() error: %v", err)
	}
	if len(agents) != 3 {
		t.Errorf("len = %d, want 3", len(agents))
	}
}

func TestAgentStore_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := agent.NewStore(db)
	proj := createTestProject(t, db)

	a, _ := store.Create(context.Background(), agent.CreateAgentParams{
		ProjectID: proj.ID, Role: "backend", Provider: "claude", ContainerID: "c-1",
	})

	_ = store.UpdateStatus(context.Background(), a.ID, "active")

	got, _ := store.GetByID(context.Background(), a.ID)
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
}

func TestAgentStore_UpdateHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := agent.NewStore(db)
	proj := createTestProject(t, db)

	a, _ := store.Create(context.Background(), agent.CreateAgentParams{
		ProjectID: proj.ID, Role: "backend", Provider: "claude", ContainerID: "c-1",
	})

	_ = store.UpdateHealth(context.Background(), a.ID, "unhealthy")

	got, _ := store.GetByID(context.Background(), a.ID)
	if got.Health != "unhealthy" {
		t.Errorf("Health = %q, want %q", got.Health, "unhealthy")
	}
}

func TestAgentStore_UpdateCurrentTask_SetAndClear(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := agent.NewStore(db)
	proj := createTestProject(t, db)

	a, _ := store.Create(context.Background(), agent.CreateAgentParams{
		ProjectID: proj.ID, Role: "backend", Provider: "claude", ContainerID: "c-1",
	})

	// Create a task to reference.
	taskStore := orchestrator.NewTaskStore(db)
	task, err := taskStore.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: proj.ID,
		Title:     "Current task test",
	})
	if err != nil {
		t.Fatalf("creating task: %v", err)
	}

	// Set a current task.
	if err := store.UpdateCurrentTask(context.Background(), a.ID, &task.ID); err != nil {
		t.Fatalf("UpdateCurrentTask(set) error: %v", err)
	}

	got, _ := store.GetByID(context.Background(), a.ID)
	if got.CurrentTaskID == nil {
		t.Fatal("CurrentTaskID should be set, got nil")
	}
	if *got.CurrentTaskID != task.ID {
		t.Errorf("CurrentTaskID = %v, want %v", *got.CurrentTaskID, task.ID)
	}

	// Clear the current task.
	if err := store.UpdateCurrentTask(context.Background(), a.ID, nil); err != nil {
		t.Fatalf("UpdateCurrentTask(clear) error: %v", err)
	}

	got, _ = store.GetByID(context.Background(), a.ID)
	if got.CurrentTaskID != nil {
		t.Errorf("CurrentTaskID should be nil after clearing, got %v", got.CurrentTaskID)
	}
}
