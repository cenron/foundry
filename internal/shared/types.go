package shared

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// ID is a UUID wrapper for all entity identifiers.
type ID = uuid.UUID

func NewID() ID {
	return uuid.New()
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing ID %q: %w", s, err)
	}
	return id, nil
}

// ProjectStatus represents the lifecycle of a project.
type ProjectStatus string

const (
	ProjectStatusDraft     ProjectStatus = "draft"
	ProjectStatusPlanning  ProjectStatus = "planning"
	ProjectStatusEstimated ProjectStatus = "estimated"
	ProjectStatusApproved  ProjectStatus = "approved"
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusPaused    ProjectStatus = "paused"
	ProjectStatusCompleted ProjectStatus = "completed"
)

func (s ProjectStatus) Value() (driver.Value, error) { return string(s), nil }

func (s *ProjectStatus) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("ProjectStatus.Scan: expected string, got %T", src)
	}
	*s = ProjectStatus(str)
	return nil
}

// TaskStatus represents the lifecycle of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusPaused     TaskStatus = "paused"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
)

func (s TaskStatus) Value() (driver.Value, error) { return string(s), nil }

func (s *TaskStatus) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("TaskStatus.Scan: expected string, got %T", src)
	}
	*s = TaskStatus(str)
	return nil
}

// AgentStatus represents the lifecycle of an agent.
type AgentStatus string

const (
	AgentStatusStarting AgentStatus = "starting"
	AgentStatusActive   AgentStatus = "active"
	AgentStatusPaused   AgentStatus = "paused"
	AgentStatusStopping AgentStatus = "stopping"
	AgentStatusStopped  AgentStatus = "stopped"
)

func (s AgentStatus) Value() (driver.Value, error) { return string(s), nil }

func (s *AgentStatus) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("AgentStatus.Scan: expected string, got %T", src)
	}
	*s = AgentStatus(str)
	return nil
}

// ApprovalStatus represents the spec approval lifecycle.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

func (s ApprovalStatus) Value() (driver.Value, error) { return string(s), nil }

func (s *ApprovalStatus) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("ApprovalStatus.Scan: expected string, got %T", src)
	}
	*s = ApprovalStatus(str)
	return nil
}

// RiskLevel classifies task risk for verification and model routing.
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

func (s RiskLevel) Value() (driver.Value, error) { return string(s), nil }

func (s *RiskLevel) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("RiskLevel.Scan: expected string, got %T", src)
	}
	*s = RiskLevel(str)
	return nil
}

// PaginationParams holds pagination query parameters.
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (p PaginationParams) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit()
}

func (p PaginationParams) Limit() int {
	if p.PageSize < 1 || p.PageSize > 100 {
		return 20
	}
	return p.PageSize
}

// PaginatedResponse wraps a list with pagination metadata.
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
}
