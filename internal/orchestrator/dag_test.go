package orchestrator_test

import (
	"testing"

	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/shared"
	"github.com/lib/pq"
)

func makeTask(id shared.ID, deps ...shared.ID) orchestrator.Task {
	depStrs := make([]string, len(deps))
	for i, d := range deps {
		depStrs[i] = d.String()
	}
	return orchestrator.Task{
		ID:        id,
		ProjectID: shared.NewID(),
		Title:     "Task " + id.String()[:8],
		Status:    "pending",
		DependsOn: pq.StringArray(depStrs),
	}
}

func TestDAGResolver_ValidateDependencies_NoCycle(t *testing.T) {
	a, b, c := shared.NewID(), shared.NewID(), shared.NewID()
	tasks := []orchestrator.Task{
		makeTask(a),
		makeTask(b, a),
		makeTask(c, b),
	}

	dag := orchestrator.NewDAGResolver(nil)
	if err := dag.ValidateDependencies(tasks); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDAGResolver_ValidateDependencies_Cycle(t *testing.T) {
	a, b, c := shared.NewID(), shared.NewID(), shared.NewID()
	tasks := []orchestrator.Task{
		makeTask(a, c), // A depends on C
		makeTask(b, a), // B depends on A
		makeTask(c, b), // C depends on B — cycle!
	}

	dag := orchestrator.NewDAGResolver(nil)
	if err := dag.ValidateDependencies(tasks); err == nil {
		t.Fatal("expected cycle detection error")
	}
}

func TestDAGResolver_ValidateDependencies_NoDeps(t *testing.T) {
	a, b := shared.NewID(), shared.NewID()
	tasks := []orchestrator.Task{
		makeTask(a),
		makeTask(b),
	}

	dag := orchestrator.NewDAGResolver(nil)
	if err := dag.ValidateDependencies(tasks); err != nil {
		t.Fatalf("expected no error for independent tasks, got: %v", err)
	}
}

func TestDAGResolver_ValidateDependencies_MultipleDeps(t *testing.T) {
	a, b, c := shared.NewID(), shared.NewID(), shared.NewID()
	tasks := []orchestrator.Task{
		makeTask(a),
		makeTask(b),
		makeTask(c, a, b), // C depends on both A and B
	}

	dag := orchestrator.NewDAGResolver(nil)
	if err := dag.ValidateDependencies(tasks); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDAGResolver_GetCriticalPath_Linear(t *testing.T) {
	a, b, c := shared.NewID(), shared.NewID(), shared.NewID()
	tasks := []orchestrator.Task{
		makeTask(a),
		makeTask(b, a),
		makeTask(c, b),
	}

	dag := orchestrator.NewDAGResolver(nil)
	path := dag.GetCriticalPath(tasks)

	if len(path) != 3 {
		t.Fatalf("critical path len = %d, want 3", len(path))
	}
	if path[0].ID != a {
		t.Errorf("path[0] = %s, want %s", path[0].ID, a)
	}
	if path[2].ID != c {
		t.Errorf("path[2] = %s, want %s", path[2].ID, c)
	}
}

func TestDAGResolver_GetCriticalPath_Diamond(t *testing.T) {
	// A → B → D
	// A → C → D
	// Critical path is 3 (A → B|C → D)
	a, b, c, d := shared.NewID(), shared.NewID(), shared.NewID(), shared.NewID()
	tasks := []orchestrator.Task{
		makeTask(a),
		makeTask(b, a),
		makeTask(c, a),
		makeTask(d, b, c),
	}

	dag := orchestrator.NewDAGResolver(nil)
	path := dag.GetCriticalPath(tasks)

	if len(path) != 3 {
		t.Fatalf("critical path len = %d, want 3", len(path))
	}
	if path[0].ID != a {
		t.Errorf("first task should be A")
	}
	if path[2].ID != d {
		t.Errorf("last task should be D")
	}
}

func TestDAGResolver_GetCriticalPath_Empty(t *testing.T) {
	dag := orchestrator.NewDAGResolver(nil)
	path := dag.GetCriticalPath(nil)
	if len(path) != 0 {
		t.Errorf("expected empty path, got %d", len(path))
	}
}

func TestDAGResolver_GetCriticalPath_Single(t *testing.T) {
	a := shared.NewID()
	tasks := []orchestrator.Task{makeTask(a)}

	dag := orchestrator.NewDAGResolver(nil)
	path := dag.GetCriticalPath(tasks)

	if len(path) != 1 {
		t.Fatalf("critical path len = %d, want 1", len(path))
	}
}
