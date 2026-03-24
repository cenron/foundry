package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/cenron/foundry/internal/agent"
)

// --- BudgetTracker mocks ---

type mockTokenStore struct {
	addErr         error
	usageByProject map[string]int
}

func (m *mockTokenStore) AddTokenUsage(_ context.Context, taskID string, tokens int) error {
	if m.addErr != nil {
		return m.addErr
	}
	return nil
}

func (m *mockTokenStore) GetProjectTokenUsage(_ context.Context, projectID string) (int, error) {
	return m.usageByProject[projectID], nil
}

type mockPublisher struct {
	published []struct {
		exchange   string
		routingKey string
		body       []byte
	}
	publishErr error
}

func (m *mockPublisher) Publish(_ context.Context, exchange, routingKey string, body []byte) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.published = append(m.published, struct {
		exchange   string
		routingKey string
		body       []byte
	}{exchange, routingKey, body})
	return nil
}

func TestTierResolver_Resolve(t *testing.T) {
	routing := json.RawMessage(`{
		"claude": {"low": "haiku", "medium": "sonnet", "high": "opus"}
	}`)

	resolver, err := agent.NewTierResolver(routing)
	if err != nil {
		t.Fatalf("NewTierResolver() error: %v", err)
	}

	tests := []struct {
		name      string
		riskLevel string
		provider  string
		roleModel string
		want      string
	}{
		{"low risk → haiku", "low", "claude", "haiku", "haiku"},
		{"medium risk → sonnet", "medium", "claude", "sonnet", "sonnet"},
		{"high risk → opus", "high", "claude", "opus", "opus"},
		{"low risk but role needs sonnet → sonnet (floor)", "low", "claude", "sonnet", "sonnet"},
		{"low risk but role needs opus → opus (floor)", "low", "claude", "opus", "opus"},
		{"unknown provider → role default", "low", "gemini", "sonnet", "sonnet"},
		{"unknown risk level → role default", "critical", "claude", "sonnet", "sonnet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := agent.AgentDefinition{Name: "test", Model: tt.roleModel}
			got := resolver.Resolve(tt.riskLevel, tt.provider, role)
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTierResolver_InvalidJSON(t *testing.T) {
	_, err := agent.NewTierResolver(json.RawMessage(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCalculateCost_ProviderReported(t *testing.T) {
	usage := agent.TokenUsage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalCostUSD: 0.42, // Provider reported directly
	}

	cost := agent.CalculateCost(usage, "sonnet", agent.DefaultPriceTable())
	if cost != 0.42 {
		t.Errorf("cost = %f, want 0.42 (provider-reported)", cost)
	}
}

func TestCalculateCost_Calculated(t *testing.T) {
	usage := agent.TokenUsage{
		InputTokens:  10_000,
		OutputTokens: 2_000,
	}

	table := agent.DefaultPriceTable()
	cost := agent.CalculateCost(usage, "sonnet", table)

	// 10K input * 0.003/1K = 0.03
	// 2K output * 0.015/1K = 0.03
	// Total = 0.06
	expected := 0.06
	if cost < expected-0.001 || cost > expected+0.001 {
		t.Errorf("cost = %f, want ~%f", cost, expected)
	}
}

func TestCalculateCost_UnknownModel(t *testing.T) {
	usage := agent.TokenUsage{InputTokens: 1000}
	cost := agent.CalculateCost(usage, "unknown-model", agent.DefaultPriceTable())
	if cost != 0 {
		t.Errorf("cost = %f, want 0 for unknown model", cost)
	}
}

func TestNewBudgetTracker(t *testing.T) {
	store := &mockTokenStore{usageByProject: map[string]int{}}
	pub := &mockPublisher{}

	bt := agent.NewBudgetTracker(store, pub)
	if bt == nil {
		t.Fatal("NewBudgetTracker() returned nil")
	}
}

func TestBudgetTracker_RecordUsage_StoreError(t *testing.T) {
	store := &mockTokenStore{
		addErr:         errors.New("db down"),
		usageByProject: map[string]int{},
	}
	pub := &mockPublisher{}

	bt := agent.NewBudgetTracker(store, pub)
	err := bt.RecordUsage(context.Background(), "proj-1", "task-1", agent.TokenUsage{
		InputTokens: 100, OutputTokens: 50,
	})

	if err == nil {
		t.Fatal("expected error when store fails, got nil")
	}
}

func TestBudgetTracker_RecordUsage_NoThresholdBelowHalf(t *testing.T) {
	// 400K tokens out of 1M = 40% — below 50% threshold, no events.
	store := &mockTokenStore{usageByProject: map[string]int{"proj-1": 400_000}}
	pub := &mockPublisher{}

	bt := agent.NewBudgetTracker(store, pub)
	if err := bt.RecordUsage(context.Background(), "proj-1", "task-1", agent.TokenUsage{
		InputTokens: 100, OutputTokens: 50,
	}); err != nil {
		t.Fatalf("RecordUsage() error: %v", err)
	}

	if len(pub.published) != 0 {
		t.Errorf("expected 0 events at 40%% utilisation, got %d", len(pub.published))
	}
}

func TestBudgetTracker_RecordUsage_Threshold50(t *testing.T) {
	// Simulate 500K / 1M = 50% — triggers the 50% threshold event.
	store := &mockTokenStore{usageByProject: map[string]int{"proj-1": 500_000}}
	pub := &mockPublisher{}

	bt := agent.NewBudgetTracker(store, pub)
	if err := bt.RecordUsage(context.Background(), "proj-1", "task-1", agent.TokenUsage{
		InputTokens: 100, OutputTokens: 50,
	}); err != nil {
		t.Fatalf("RecordUsage() error: %v", err)
	}

	if len(pub.published) != 1 {
		t.Fatalf("expected 1 threshold event at 50%%, got %d", len(pub.published))
	}
}

func TestBudgetTracker_RecordUsage_Threshold75(t *testing.T) {
	// 750K / 1M = 75% — fires both 50% and 75% thresholds on first call.
	store := &mockTokenStore{usageByProject: map[string]int{"proj-1": 750_000}}
	pub := &mockPublisher{}

	bt := agent.NewBudgetTracker(store, pub)
	if err := bt.RecordUsage(context.Background(), "proj-1", "task-1", agent.TokenUsage{
		InputTokens: 100, OutputTokens: 50,
	}); err != nil {
		t.Fatalf("RecordUsage() error: %v", err)
	}

	if len(pub.published) != 2 {
		t.Fatalf("expected 2 threshold events at 75%%, got %d", len(pub.published))
	}
}

func TestBudgetTracker_RecordUsage_Threshold90(t *testing.T) {
	// 900K / 1M = 90% — fires 50%, 75%, and 90% thresholds on first call.
	store := &mockTokenStore{usageByProject: map[string]int{"proj-1": 900_000}}
	pub := &mockPublisher{}

	bt := agent.NewBudgetTracker(store, pub)
	if err := bt.RecordUsage(context.Background(), "proj-1", "task-1", agent.TokenUsage{
		InputTokens: 100, OutputTokens: 50,
	}); err != nil {
		t.Fatalf("RecordUsage() error: %v", err)
	}

	if len(pub.published) != 3 {
		t.Fatalf("expected 3 threshold events at 90%%, got %d", len(pub.published))
	}
}

func TestBudgetTracker_RecordUsage_NoDuplicateThresholdEvents(t *testing.T) {
	// Call RecordUsage twice at 90% — the second call should produce no new events
	// because each threshold is only emitted once per project.
	store := &mockTokenStore{usageByProject: map[string]int{"proj-1": 900_000}}
	pub := &mockPublisher{}

	bt := agent.NewBudgetTracker(store, pub)

	for i := 0; i < 2; i++ {
		if err := bt.RecordUsage(context.Background(), "proj-1", "task-1", agent.TokenUsage{
			InputTokens: 100, OutputTokens: 50,
		}); err != nil {
			t.Fatalf("RecordUsage() call %d error: %v", i, err)
		}
	}

	// Still just 3 events total — no duplicates on second call.
	if len(pub.published) != 3 {
		t.Errorf("expected 3 unique threshold events, got %d", len(pub.published))
	}
}

func TestBudgetTracker_RecordUsage_PublishError_DoesNotFailRecordUsage(t *testing.T) {
	// Publish failures are logged but do not surface as errors from RecordUsage.
	store := &mockTokenStore{usageByProject: map[string]int{"proj-1": 900_000}}
	pub := &mockPublisher{publishErr: errors.New("broker offline")}

	bt := agent.NewBudgetTracker(store, pub)
	err := bt.RecordUsage(context.Background(), "proj-1", "task-1", agent.TokenUsage{
		InputTokens: 100, OutputTokens: 50,
	})

	if err != nil {
		t.Errorf("RecordUsage() should succeed even when publish fails, got: %v", err)
	}
}
