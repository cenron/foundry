package shared_test

import (
	"testing"

	"github.com/cenron/foundry/internal/shared"
)

func TestNewID(t *testing.T) {
	id := shared.NewID()
	if id.String() == "" {
		t.Error("NewID() returned empty ID")
	}

	// Each call should produce a unique ID.
	id2 := shared.NewID()
	if id == id2 {
		t.Error("NewID() returned the same ID twice")
	}
}

func TestParseID_Valid(t *testing.T) {
	original := shared.NewID()
	parsed, err := shared.ParseID(original.String())
	if err != nil {
		t.Fatalf("ParseID() error: %v", err)
	}
	if parsed != original {
		t.Errorf("parsed ID %v != original %v", parsed, original)
	}
}

func TestParseID_Invalid(t *testing.T) {
	_, err := shared.ParseID("not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
}

func TestNotFoundError_Error(t *testing.T) {
	err := &shared.NotFoundError{Resource: "project", ID: "abc-123"}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &shared.ValidationError{Field: "name", Message: "required"}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestConflictError_Error(t *testing.T) {
	err := &shared.ConflictError{Resource: "spec", Message: "already approved"}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		want     int
	}{
		{"page 1", 1, 10, 0},
		{"page 2", 2, 10, 10},
		{"page 3", 3, 20, 40},
		{"page below 1 → 0", 0, 10, 0},
		{"negative page → 0", -1, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := shared.PaginationParams{Page: tt.page, PageSize: tt.pageSize}
			if got := p.Offset(); got != tt.want {
				t.Errorf("Offset() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPaginationParams_Limit(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		want     int
	}{
		{"normal size", 10, 10},
		{"max size", 100, 100},
		{"too large → 20 default", 101, 20},
		{"zero → 20 default", 0, 20},
		{"negative → 20 default", -5, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := shared.PaginationParams{PageSize: tt.pageSize}
			if got := p.Limit(); got != tt.want {
				t.Errorf("Limit() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestProjectStatus_ValueAndScan(t *testing.T) {
	original := shared.ProjectStatusActive
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if val != "active" {
		t.Errorf("Value() = %q, want %q", val, "active")
	}

	var scanned shared.ProjectStatus
	if err := scanned.Scan("approved"); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if scanned != shared.ProjectStatusApproved {
		t.Errorf("Scan() = %q, want %q", scanned, shared.ProjectStatusApproved)
	}
}

func TestProjectStatus_Scan_InvalidType(t *testing.T) {
	var s shared.ProjectStatus
	if err := s.Scan(42); err == nil {
		t.Fatal("expected error scanning non-string type")
	}
}

func TestTaskStatus_ValueAndScan(t *testing.T) {
	original := shared.TaskStatusDone
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if val != "done" {
		t.Errorf("Value() = %q, want %q", val, "done")
	}

	var scanned shared.TaskStatus
	if err := scanned.Scan("pending"); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if scanned != shared.TaskStatusPending {
		t.Errorf("Scan() = %q, want %q", scanned, shared.TaskStatusPending)
	}
}

func TestTaskStatus_Scan_InvalidType(t *testing.T) {
	var s shared.TaskStatus
	if err := s.Scan(true); err == nil {
		t.Fatal("expected error scanning non-string type")
	}
}

func TestAgentStatus_ValueAndScan(t *testing.T) {
	original := shared.AgentStatusActive
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if val != "active" {
		t.Errorf("Value() = %q, want %q", val, "active")
	}

	var scanned shared.AgentStatus
	if err := scanned.Scan("stopped"); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if scanned != shared.AgentStatusStopped {
		t.Errorf("Scan() = %q, want %q", scanned, shared.AgentStatusStopped)
	}
}

func TestAgentStatus_Scan_InvalidType(t *testing.T) {
	var s shared.AgentStatus
	if err := s.Scan(nil); err == nil {
		t.Fatal("expected error scanning non-string type")
	}
}

func TestApprovalStatus_ValueAndScan(t *testing.T) {
	original := shared.ApprovalStatusApproved
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if val != "approved" {
		t.Errorf("Value() = %q, want %q", val, "approved")
	}

	var scanned shared.ApprovalStatus
	if err := scanned.Scan("rejected"); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if scanned != shared.ApprovalStatusRejected {
		t.Errorf("Scan() = %q, want %q", scanned, shared.ApprovalStatusRejected)
	}
}

func TestApprovalStatus_Scan_InvalidType(t *testing.T) {
	var s shared.ApprovalStatus
	if err := s.Scan(3.14); err == nil {
		t.Fatal("expected error scanning non-string type")
	}
}

func TestRiskLevel_ValueAndScan(t *testing.T) {
	original := shared.RiskLevelHigh
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if val != "high" {
		t.Errorf("Value() = %q, want %q", val, "high")
	}

	var scanned shared.RiskLevel
	if err := scanned.Scan("low"); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if scanned != shared.RiskLevelLow {
		t.Errorf("Scan() = %q, want %q", scanned, shared.RiskLevelLow)
	}
}

func TestRiskLevel_Scan_InvalidType(t *testing.T) {
	var s shared.RiskLevel
	if err := s.Scan([]byte("medium")); err == nil {
		t.Fatal("expected error scanning non-string type")
	}
}
