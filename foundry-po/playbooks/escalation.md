# Escalation session

An agent hit a blocker and needs help.

## On startup

1. Read session context for task_id, agent_role, and the blocker description
2. Read `projects/<name>/plan.md` for task context
3. Read the agent's recent output/logs if available

## Your job

Diagnose the problem. Options:
- Provide guidance and send the agent back to retry
- Escalate the task's risk level (triggers model upgrade on next session)
- Reassign the task to a different agent role
- Break the task into smaller subtasks
- Flag to the user if it requires human judgment

## Outputs

- Update `plan.md` with your decision
- Write guidance to `artifacts/` if sending the agent back
- Update task risk level if escalating
- Write to `memory/` if this is a lesson for future projects
