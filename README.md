# Foundry

Spec-driven AI dev platform — orchestrate teams of Claude Code agents with risk-based verification, cost-aware model routing, and a living execution narrative.

Foundry turns specifications into working software. You plan with an AI Product Owner, approve the spec, and Foundry spins up a team of specialized agents — backend, frontend, QA — working in parallel inside Docker containers with git worktrees for branch isolation. A Go control plane manages infrastructure while the PO handles judgment calls: decomposing work, reviewing output, and maintaining a living spec that documents every decision made during execution. Risk classification drives everything — verification depth, model selection (Haiku for boilerplate, Opus for auth), and PO attention — so you get quality where it matters and save tokens where it doesn't.

## Setup

```bash
# Start dev infrastructure (Postgres, Redis, RabbitMQ)
make docker-up

# Run database migrations
make migrate-up

# Start the backend (with hot reload)
make run

# Start the frontend
cd web && npm install && npm run dev
```

## Development

```bash
# Run all tests
make test-all

# Backend only
make test
make lint

# Frontend only
cd web && npm test
cd web && npm run lint
cd web && npx playwright test
```
