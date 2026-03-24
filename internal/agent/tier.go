package agent

import (
	"encoding/json"
	"fmt"
)

// ModelRouting maps risk levels to model tiers per provider.
// Example: {"claude": {"low": "haiku", "medium": "sonnet", "high": "opus"}}
type ModelRouting map[string]map[string]string

// TierResolver resolves the concrete model tier for a task based on risk level
// and agent role, using the project's risk profile.
type TierResolver struct {
	routing ModelRouting
}

func NewTierResolver(routingJSON json.RawMessage) (*TierResolver, error) {
	var routing ModelRouting
	if err := json.Unmarshal(routingJSON, &routing); err != nil {
		return nil, fmt.Errorf("parsing model routing: %w", err)
	}
	return &TierResolver{routing: routing}, nil
}

// Resolve returns the model tier for a given risk level, provider, and agent role.
// It respects the role's minimum model floor — e.g., a code-reviewer should never
// get downgraded below sonnet even on low-risk tasks.
func (r *TierResolver) Resolve(riskLevel, provider string, role AgentDefinition) string {
	providerRouting, ok := r.routing[provider]
	if !ok {
		return role.Model // fallback to role default
	}

	resolved, ok := providerRouting[riskLevel]
	if !ok {
		return role.Model
	}

	// Enforce minimum model floor from the role definition
	if tierRank(resolved) < tierRank(role.Model) {
		return role.Model
	}

	return resolved
}

// tierRank returns the capability rank of a model tier.
// Higher rank = more capable model.
func tierRank(tier string) int {
	ranks := map[string]int{
		"haiku":  1,
		"sonnet": 2,
		"opus":   3,
	}
	if r, ok := ranks[tier]; ok {
		return r
	}
	return 0
}
