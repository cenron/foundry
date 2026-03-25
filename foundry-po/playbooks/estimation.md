# Estimation session

The spec has been approved. Generate the execution plan.

## On startup

1. Read `projects/<name>/approved_spec.md`
2. Read `projects/<name>/decisions/`
3. Load language and framework files for the tech stack

## Your job

1. Decompose the spec into phases and tasks
2. Identify dependencies between tasks (what blocks what)
3. Classify each task's risk level (low/medium/high)
4. Determine team composition — which agent roles are needed
5. Estimate token cost per phase based on task complexity and model tier (low risk = haiku, medium = sonnet, high = opus)
6. Identify which tasks can be parallelized within each phase

## Plan format

Use checkbox syntax for tasks with metadata:

```markdown
## Phase 1: <name>
Status: pending

- [ ] **Task 1.1: <title>** [risk: medium] [role: backend-developer] [branch: feat/<name>]
  Description of the task.
  Depends on: none

- [ ] **Task 1.2: <title>** [risk: low] [role: frontend-developer] [branch: feat/<name>]
  Description of the task.
  Depends on: Task 1.1
```

## Outputs

Write these files — the control plane parses them:

**`plan.md`** — phased execution plan with tasks, dependencies, risk levels, and role assignments.

**`team.json`**:
```json
{
  "roles": ["backend-developer", "frontend-developer"],
  "agent_count": 2,
  "token_estimate": 150000,
  "cost_estimate_usd": 12.50
}
```

**`estimate.json`**:
```json
{
  "phases": [
    {"name": "Phase 1", "tasks": 5, "token_estimate": 50000, "cost_estimate_usd": 4.00}
  ],
  "total_tokens": 150000,
  "total_cost_usd": 12.50
}
```

Tell the user the plan is ready for review in the UI.
