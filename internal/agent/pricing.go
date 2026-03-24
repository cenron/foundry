package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// TokenUsage tracks token consumption for a single interaction.
type TokenUsage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CachedTokens int     `json:"cached_tokens"`
	TotalCostUSD float64 `json:"total_cost_usd"` // from provider if available
}

// ModelPricing defines per-1K-token costs for a model.
type ModelPricing struct {
	InputPer1K  float64 `json:"input_per_1k"`
	OutputPer1K float64 `json:"output_per_1k"`
	CachedPer1K float64 `json:"cached_per_1k"`
}

// PriceTable maps model names to their pricing.
type PriceTable map[string]ModelPricing

// DefaultPriceTable returns pricing for Claude models (as of early 2026).
func DefaultPriceTable() PriceTable {
	return PriceTable{
		"haiku":  {InputPer1K: 0.00025, OutputPer1K: 0.00125, CachedPer1K: 0.0000625},
		"sonnet": {InputPer1K: 0.003, OutputPer1K: 0.015, CachedPer1K: 0.00075},
		"opus":   {InputPer1K: 0.015, OutputPer1K: 0.075, CachedPer1K: 0.00375},
	}
}

// CalculateCost computes the cost from token usage.
// If the provider reported total_cost_usd directly, that takes precedence.
func CalculateCost(usage TokenUsage, model string, table PriceTable) float64 {
	if usage.TotalCostUSD > 0 {
		return usage.TotalCostUSD
	}

	pricing, ok := table[model]
	if !ok {
		return 0
	}

	input := float64(usage.InputTokens) / 1000 * pricing.InputPer1K
	output := float64(usage.OutputTokens) / 1000 * pricing.OutputPer1K
	cached := float64(usage.CachedTokens) / 1000 * pricing.CachedPer1K

	return input + output + cached
}

// BudgetStatus tracks project-level budget utilization.
type BudgetStatus struct {
	ProjectID      string  `json:"project_id"`
	TotalBudgetUSD float64 `json:"total_budget_usd"`
	SpentUSD       float64 `json:"spent_usd"`
	Utilization    float64 `json:"utilization"` // 0.0–1.0
}

// TokenUsageUpdater persists token usage to the data layer.
type TokenUsageUpdater interface {
	AddTokenUsage(ctx context.Context, taskID string, tokens int) error
	GetProjectTokenUsage(ctx context.Context, projectID string) (int, error)
}

// BudgetEventPublisher publishes budget threshold events.
type BudgetEventPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

type BudgetTracker struct {
	store     TokenUsageUpdater
	publisher BudgetEventPublisher

	// emittedThresholds tracks the highest threshold emitted per project
	// to prevent duplicate events on every token update.
	emittedThresholds map[string]float64
}

func NewBudgetTracker(store TokenUsageUpdater, publisher BudgetEventPublisher) *BudgetTracker {
	return &BudgetTracker{
		store:             store,
		publisher:         publisher,
		emittedThresholds: make(map[string]float64),
	}
}

// RecordUsage records token usage and checks budget thresholds.
func (bt *BudgetTracker) RecordUsage(ctx context.Context, projectID, taskID string, usage TokenUsage) error {
	totalTokens := usage.InputTokens + usage.OutputTokens
	if err := bt.store.AddTokenUsage(ctx, taskID, totalTokens); err != nil {
		return fmt.Errorf("recording token usage: %w", err)
	}

	bt.checkThresholds(ctx, projectID)

	return nil
}

func (bt *BudgetTracker) checkThresholds(ctx context.Context, projectID string) {
	thresholds := []float64{0.50, 0.75, 0.90}

	totalTokens, err := bt.store.GetProjectTokenUsage(ctx, projectID)
	if err != nil {
		log.Printf("budget tracker: getting project usage: %v", err)
		return
	}

	// Simple threshold check based on a default budget of 1M tokens.
	// In production this would read from the project's spec.token_estimate.
	budgetTokens := 1_000_000
	utilization := float64(totalTokens) / float64(budgetTokens)

	for _, threshold := range thresholds {
		if utilization < threshold {
			continue
		}
		// Only emit each threshold once — skip if already emitted.
		if bt.emittedThresholds[projectID] >= threshold {
			continue
		}
		bt.emittedThresholds[projectID] = threshold
		bt.publishThresholdEvent(ctx, projectID, threshold, utilization)
	}
}

func (bt *BudgetTracker) publishThresholdEvent(ctx context.Context, projectID string, threshold, utilization float64) {
	payload, err := json.Marshal(map[string]interface{}{
		"project_id":  projectID,
		"type":        "budget_threshold",
		"threshold":   threshold,
		"utilization": utilization,
	})
	if err != nil {
		log.Printf("budget tracker: marshaling threshold event: %v", err)
		return
	}

	routingKey := fmt.Sprintf("events.%s.budget_threshold", projectID)
	if err := bt.publisher.Publish(ctx, "foundry.events", routingKey, payload); err != nil {
		log.Printf("budget tracker: publishing threshold event: %v", err)
	}
}
