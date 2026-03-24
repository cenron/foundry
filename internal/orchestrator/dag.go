package orchestrator

import (
	"context"
	"fmt"

	"github.com/cenron/foundry/internal/shared"
)

// UnblockedTaskFinder queries for tasks ready to execute.
type UnblockedTaskFinder interface {
	GetUnblockedTasks(ctx context.Context, projectID shared.ID) ([]Task, error)
}

type DAGResolver struct {
	store UnblockedTaskFinder
}

func NewDAGResolver(store UnblockedTaskFinder) *DAGResolver {
	return &DAGResolver{store: store}
}

// GetUnblockedTasks delegates to the store's dependency-aware query.
func (d *DAGResolver) GetUnblockedTasks(ctx context.Context, projectID shared.ID) ([]Task, error) {
	return d.store.GetUnblockedTasks(ctx, projectID)
}

// ValidateDependencies checks for cycles using topological sort (Kahn's algorithm).
func (d *DAGResolver) ValidateDependencies(tasks []Task) error {
	taskMap := buildTaskIndex(tasks)

	// Build in-degree count and adjacency list
	inDegree := make(map[string]int)
	dependents := make(map[string][]string) // depID → tasks that depend on it

	for _, t := range tasks {
		id := t.ID.String()
		if _, ok := inDegree[id]; !ok {
			inDegree[id] = 0
		}

		for _, depID := range t.DependsOn {
			inDegree[id]++
			dependents[depID] = append(dependents[depID], id)
		}
	}

	// Start with tasks that have no dependencies
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	visited := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		visited++

		for _, depID := range dependents[current] {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				queue = append(queue, depID)
			}
		}
	}

	if visited != len(taskMap) {
		return fmt.Errorf("circular dependency detected: %d of %d tasks could not be resolved", len(taskMap)-visited, len(taskMap))
	}

	return nil
}

// GetCriticalPath returns the longest dependency chain through the task graph.
func (d *DAGResolver) GetCriticalPath(tasks []Task) []Task {
	taskMap := buildTaskIndex(tasks)

	// Memoized longest path from each task
	memo := make(map[string]int)
	parent := make(map[string]string) // child → parent on critical path

	var longestFrom func(id string) int
	longestFrom = func(id string) int {
		if v, ok := memo[id]; ok {
			return v
		}

		t, ok := taskMap[id]
		if !ok {
			return 0
		}

		best := 0
		bestParent := ""
		for _, depID := range t.DependsOn {
			depLen := longestFrom(depID)
			if depLen > best {
				best = depLen
				bestParent = depID
			}
		}

		memo[id] = best + 1
		if bestParent != "" {
			parent[id] = bestParent
		}
		return best + 1
	}

	// Find the task with the longest path
	var endID string
	maxLen := 0
	for _, t := range tasks {
		id := t.ID.String()
		pathLen := longestFrom(id)
		if pathLen > maxLen {
			maxLen = pathLen
			endID = id
		}
	}

	if endID == "" {
		return nil
	}

	// Walk back from the end to reconstruct the path
	var path []Task
	for current := endID; current != ""; current = parent[current] {
		if t, ok := taskMap[current]; ok {
			path = append(path, t)
		}
	}

	// Reverse to get start → end order
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

func buildTaskIndex(tasks []Task) map[string]Task {
	m := make(map[string]Task, len(tasks))
	for _, t := range tasks {
		m[t.ID.String()] = t
	}
	return m
}
