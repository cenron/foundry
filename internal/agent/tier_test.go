package agent_test

import (
	"encoding/json"
	"testing"

	"github.com/cenron/foundry/internal/agent"
)

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
