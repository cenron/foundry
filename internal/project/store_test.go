package project_test

import (
	"context"
	"os"
	"testing"

	"github.com/cenron/foundry/internal/database"
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
		_, _ = db.Exec("DELETE FROM specs")
		_, _ = db.Exec("DELETE FROM projects")
		_ = db.Close()
	})

	return db
}

func TestProjectStore_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := project.NewStore(db)

	p, err := store.Create(context.Background(), project.CreateProjectParams{
		Name:            "Test Project",
		Description:     "A test project",
		RepoURL:         "https://github.com/test/repo",
		TeamComposition: []string{"backend-developer", "frontend-developer"},
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if p.Name != "Test Project" {
		t.Errorf("Name = %q, want %q", p.Name, "Test Project")
	}
	if p.Status != "draft" {
		t.Errorf("Status = %q, want %q", p.Status, "draft")
	}
}

func TestProjectStore_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := project.NewStore(db)

	created, err := store.Create(context.Background(), project.CreateProjectParams{
		Name: "Get Test",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := store.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %v, want %v", got.ID, created.ID)
	}
}

func TestProjectStore_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := project.NewStore(db)

	for i := 0; i < 3; i++ {
		_, err := store.Create(context.Background(), project.CreateProjectParams{
			Name: "List Test",
		})
		if err != nil {
			t.Fatalf("Create() error: %v", err)
		}
	}

	projects, total, err := store.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if total < 3 {
		t.Errorf("total = %d, want >= 3", total)
	}
	if len(projects) < 3 {
		t.Errorf("len = %d, want >= 3", len(projects))
	}
}

func TestProjectStore_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	store := project.NewStore(db)

	p, _ := store.Create(context.Background(), project.CreateProjectParams{Name: "Status Test"})

	err := store.UpdateStatus(context.Background(), p.ID, "active")
	if err != nil {
		t.Fatalf("UpdateStatus() error: %v", err)
	}

	got, _ := store.GetByID(context.Background(), p.ID)
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
}

func TestSpecStore_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	projectStore := project.NewStore(db)
	specStore := project.NewSpecStore(db)

	p, _ := projectStore.Create(context.Background(), project.CreateProjectParams{Name: "Spec Test"})

	spec, err := specStore.Create(context.Background(), project.CreateSpecParams{
		ProjectID:       p.ID,
		ApprovedContent: "# Spec Content",
		TokenEstimate:   50000,
		AgentCount:      3,
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if spec.ApprovalStatus != "pending" {
		t.Errorf("ApprovalStatus = %q, want %q", spec.ApprovalStatus, "pending")
	}

	got, err := specStore.GetByProjectID(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("GetByProjectID() error: %v", err)
	}

	if got.ID != spec.ID {
		t.Errorf("ID = %v, want %v", got.ID, spec.ID)
	}
}

func TestSpecStore_UpdateApproval(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	projectStore := project.NewStore(db)
	specStore := project.NewSpecStore(db)

	p, _ := projectStore.Create(context.Background(), project.CreateProjectParams{Name: "Approval Test"})
	spec, _ := specStore.Create(context.Background(), project.CreateSpecParams{
		ProjectID:       p.ID,
		ApprovedContent: "# Spec",
	})

	err := specStore.UpdateApproval(context.Background(), spec.ID, "approved")
	if err != nil {
		t.Fatalf("UpdateApproval() error: %v", err)
	}

	got, _ := specStore.GetByProjectID(context.Background(), p.ID)
	if got.ApprovalStatus != "approved" {
		t.Errorf("ApprovalStatus = %q, want %q", got.ApprovalStatus, "approved")
	}
	if got.ApprovedAt == nil {
		t.Error("ApprovedAt should be set after approval")
	}
}
