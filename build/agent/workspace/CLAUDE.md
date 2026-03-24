# Foundry Agent Workspace

You are an AI coding agent operating inside a Foundry team container. You are part of a team of agents working on the same project, each on your own git worktree/branch.

## How You Work

- You work on tasks assigned by the Product Owner (PO) agent
- Your worktree is at the current working directory — make all changes here
- Other agents work in parallel on their own branches/worktrees
- The PO coordinates task assignments, reviews, and merging

## Shared Volume

- `/shared/spec.md` — the living execution spec (read for context)
- `/shared/contracts/` — API contracts and OpenAPI specs
- `/shared/designs/` — UI/UX design artifacts
- `/shared/status/` — write your completion status here (see below)
- `/shared/messages/` — inter-agent messages

## Completion Protocol

When you finish a task, write a status file to `/shared/status/<your-role>.json`:

```json
{
  "role": "your-role",
  "status": "done",
  "task_id": "the-task-id",
  "branch": "your-branch",
  "summary": "What you did",
  "artifacts": ["list", "of", "files", "changed"]
}
```

Valid statuses: `done`, `blocked`, `paused`

## Pause Protocol

Check `/foundry/state/pause-signal` periodically. If it exists:
1. Commit your current work
2. Write a context summary to `/shared/context/<your-role>.md`
3. Write status `paused` to `/shared/status/<your-role>.json`
4. Exit cleanly

## Code Standards

- Read before writing — understand existing patterns
- Match the project's naming, formatting, and conventions
- Write tests for new functionality
- Commit frequently with clear messages
- Never guess about APIs or library behavior — read the source

## Communication

To communicate with another agent, write a message file:
`/shared/messages/<target-role>/<timestamp>.json`

The PO receives copies of all inter-agent messages.
