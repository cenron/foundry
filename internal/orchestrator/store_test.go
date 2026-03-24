package orchestrator_test

import (
	"context"
	"os"
	"testing"

	"github.com/cenron/foundry/internal/database"
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
		_, _ = db.Exec("DELETE FROM tasks")
		_, _ = db.Exec("DELETE FROM projects")
		_ = db.Close()
	})

	return db
}

func createTestProject(t *testing.T, db *sqlx.DB) *project.Project {
	t.Helper()
	store := project.NewStore(db)
	p, err := store.Create(context.Background(), project.CreateProjectParams{Name: "Task Test Project"})
	if err != nil {
		t.Fatalf("creating test project: %v", err)
	}
	return p
}

func TestTaskStore_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := orchestrator.NewTaskStore(db)
	proj := createTestProject(t, db)

	task, err := store.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID:    proj.ID,
		Title:        "Implement auth",
		Description:  "Build authentication module",
		RiskLevel:    "high",
		AssignedRole: "backend-developer",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if task.Title != "Implement auth" {
		t.Errorf("Title = %q, want %q", task.Title, "Implement auth")
	}
	if task.RiskLevel != "high" {
		t.Errorf("RiskLevel = %q, want %q", task.RiskLevel, "high")
	}
	if task.Status != "pending" {
		t.Errorf("Status = %q, want %q", task.Status, "pending")
	}
}

func TestTaskStore_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := orchestrator.NewTaskStore(db)
	proj := createTestProject(t, db)

	created, _ := store.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: proj.ID,
		Title:     "Get Test",
	})

	got, err := store.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID mismatch")
	}
}

func TestTaskStore_ListByProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := orchestrator.NewTaskStore(db)
	proj := createTestProject(t, db)

	for i := 0; i < 3; i++ {
		_, _ = store.Create(context.Background(), orchestrator.CreateTaskParams{
			ProjectID: proj.ID,
			Title:     "List task",
		})
	}

	tasks, err := store.ListByProject(context.Background(), proj.ID)
	if err != nil {
		t.Fatalf("ListByProject() error: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("len = %d, want 3", len(tasks))
	}
}

func TestTaskStore_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := orchestrator.NewTaskStore(db)
	proj := createTestProject(t, db)

	task, _ := store.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: proj.ID,
		Title:     "Status test",
	})

	_ = store.UpdateStatus(context.Background(), task.ID, "in_progress")

	got, _ := store.GetByID(context.Background(), task.ID)
	if got.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", got.Status, "in_progress")
	}
}

func TestTaskStore_GetUnblockedTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := orchestrator.NewTaskStore(db)
	proj := createTestProject(t, db)

	// Task A: no dependencies
	taskA, _ := store.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: proj.ID,
		Title:     "Task A (no deps)",
	})

	// Task B: depends on A
	taskB, _ := store.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: proj.ID,
		Title:     "Task B (depends on A)",
		DependsOn: []shared.ID{taskA.ID},
	})

	// Task C: depends on B
	_, _ = store.Create(context.Background(), orchestrator.CreateTaskParams{
		ProjectID: proj.ID,
		Title:     "Task C (depends on B)",
		DependsOn: []shared.ID{taskB.ID},
	})

	// Initially only Task A is unblocked
	unblocked, err := store.GetUnblockedTasks(context.Background(), proj.ID)
	if err != nil {
		t.Fatalf("GetUnblockedTasks() error: %v", err)
	}
	if len(unblocked) != 1 {
		t.Fatalf("expected 1 unblocked task, got %d", len(unblocked))
	}
	if unblocked[0].ID != taskA.ID {
		t.Errorf("expected Task A to be unblocked")
	}

	// Mark A as done — now B should be unblocked
	_ = store.UpdateStatus(context.Background(), taskA.ID, "done")

	unblocked, err = store.GetUnblockedTasks(context.Background(), proj.ID)
	if err != nil {
		t.Fatalf("GetUnblockedTasks() after A done: %v", err)
	}
	if len(unblocked) != 1 {
		t.Fatalf("expected 1 unblocked task after A done, got %d", len(unblocked))
	}
	if unblocked[0].ID != taskB.ID {
		t.Errorf("expected Task B to be unblocked after A done")
	}
}
