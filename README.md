# foundry

Spec-driven AI development platform that orchestrates teams of Claude Code agents to build software from specifications.

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
