# foundry — Project Conventions

> Read `docs/design-principles.md` at the start of every task. It contains the core engineering philosophy and locked sections that must not be modified unless explicitly asked.

## What This Is
Spec-driven AI development platform that orchestrates teams of Claude Code agents to build software from specifications.

## Tech Stack

### Go Conventions

- Use **air** for hot reloading during development (`air` watches for file changes and rebuilds automatically).
- Set up a `.air.toml` config at the project root to configure build commands, watched directories, and excluded paths.
- Use `go-chi/chi` for HTTP routing. Lightweight, `net/http` compatible, composable middleware.
- Use `jmoiron/sqlx` for database access. Hand-written SQL, no ORM. `sqlx.Get`/`sqlx.Select` for struct scanning, `db.NamedExec` for named parameters.
- Use `github.com/swaggo/swag` for Swagger annotations on every handler. Annotations are mandatory — the frontend generates its API client from `swagger.json`.
- Use `github.com/swaggo/http-swagger/v2` to serve Swagger UI.
- Run `make swagger` to generate `api/swagger/swagger.json` from annotations.
- Run `make generate` to regenerate both swagger spec and frontend API client.
- Use `gorilla/websocket` for WebSocket support.
- Use Docker SDK for Go (`github.com/docker/docker/client`) for container management.
- Use `rabbitmq/amqp091-go` for RabbitMQ (import alias: `amqp`).
- Use `go-redis/v9` for Redis.

When debugging library internals in Go, find the source in the module cache:
```bash
find $(go env GOMODCACHE) -path "*<module>*" -name "*.go" | xargs grep -l "<symbol>"
```

### Node/TypeScript Conventions

- Use the project's package manager (`npm`, `pnpm`, or `yarn`) consistently — don't mix.
- `npm run dev` / `pnpm dev` — start the dev server
- `npm test` / `pnpm test` — run tests
- `npm run lint` / `pnpm lint` — run linter

When debugging library internals in Node, find the source in node_modules:
```bash
find node_modules -name "*.js" -path "*<package>*" | head -20
```

## Commands

```bash
# Backend
make build              # Build Go binary
make run                # Run with air (hot reload)
make test               # Run Go tests
make lint               # Run golangci-lint
make docker-up          # Start Postgres, Redis, RabbitMQ
make docker-down        # Stop dev infrastructure
make migrate-up         # Run database migrations
make migrate-down       # Rollback last migration

# Swagger / API codegen
make swagger            # Generate swagger.json from Go annotations
make web-generate-api   # Generate frontend API client from swagger
make generate           # Both: swagger + web-generate-api

# Frontend
cd web && npm run dev   # Start Vite dev server
cd web && npm test      # Run Vitest
cd web && npm run lint  # Run ESLint
cd web && npx playwright test  # Run Playwright E2E tests

# Full suite
make test-all           # Backend tests + frontend tests + lint + e2e
```

## Architecture

See `docs/specs/design-specification.md` for the full architecture. Key points:

- **Hybrid orchestration**: Go control plane (infrastructure) + Claude Code PO agent (intelligence)
- **One team container per project**: all agents share it, git worktrees for branch isolation
- **PO runs locally**: not containerized, stateless sessions with playbook-based behavior
- **Risk-based everything**: risk classification drives verification depth, model tier, and PO attention

```
foundry/
├── cmd/foundry/                    # Entry point, composition root
├── internal/
│   ├── config/                     # Configuration loading
│   ├── database/                   # Postgres connection (sqlx), migrations
│   ├── cache/                      # Redis client, response cache
│   ├── broker/                     # RabbitMQ client, exchanges, message routing
│   ├── container/                  # Docker team container management, health monitor
│   ├── orchestrator/               # DAG resolver, task state machine, project starter
│   ├── agent/                      # Agent provider interface, registry, tier resolution, library loader
│   ├── po/                         # PO session manager, context builder
│   ├── project/                    # Project CRUD
│   ├── spec/                       # Spec CRUD, mutation tracking
│   ├── verification/               # Verification orchestration, risk classification
│   ├── event/                      # Event logging, routing
│   ├── api/                        # HTTP handlers (chi), WebSocket, middleware
│   └── shared/                     # Shared types, errors, helpers
├── web/                            # React frontend (Vite + TypeScript + shadcn/ui)
├── build/
│   ├── docker-compose.yml          # Dev infrastructure (Postgres, Redis, RabbitMQ)
│   ├── Dockerfile                  # Foundry backend image
│   └── agent/                      # Team container image
├── foundry-po/                     # PO workspace source (deploys to ~/foundry/)
│   ├── CLAUDE.md                   # Base PO instructions
│   └── playbooks/                  # Session-type playbooks (6 files)
├── migrations/                     # SQL migration files
├── docs/
│   ├── design-principles.md        # Core engineering philosophy
│   ├── specs/                      # Design specs and implementation plan
│   └── decisions/                  # Architecture decision records
├── tests/e2e/                      # End-to-end smoke tests
└── scripts/                        # Dev scripts
```

## Phased Build Workflow

> **This is how work gets done on this project.** Every module follows this cycle. No exceptions.

### The cycle

```
Phase start
  └─ For each task in the phase:
       1. Write failing test
       2. Implement minimum code to pass
       3. Run module tests → fix until green
       4. Commit
  └─ Phase gate (ALL must pass before PR):
       ├── go test ./...                    (all Go tests)
       ├── golangci-lint run                (linting)
       ├── cd web && npm test               (frontend unit tests)
       ├── cd web && npm run lint           (frontend linting)
       ├── cd web && npx playwright test    (E2E browser tests — if phase has UI)
       └── Loop until ALL green
  └─ Create PR for this phase
  └─ STOP and wait for review instructions
```

### Rules

1. **Never skip the phase gate.** Every test suite runs at the end of every phase. If any fail, fix them before the PR.
2. **80% code coverage minimum — non-negotiable.** Every phase must maintain at least 80% statement coverage for Go (`go test -cover`) and 70% for frontend (`vitest --coverage`). Measure before the PR. If coverage drops below threshold, add tests before proceeding. No exceptions.
3. **UI phases use Playwright.** Any phase that adds or modifies UI components must include comprehensive Playwright E2E tests. E2E tests are part of the phase, not an afterthought. Write them alongside the components.
4. **One PR per phase.** Each phase is a logical unit of work with its own branch (`phase/<name>` or `feat/<name>`). One PR per phase, opened only when the phase gate passes.
5. **Stop after the PR.** Do not start the next phase until instructed. The PR may have review feedback that needs addressing first.
6. **Test between steps.** Don't batch up implementation. Write test → implement → verify → commit. Tight loops catch issues early.
7. **Fix forward, don't skip.** If a test fails at the phase gate, fix the issue. Don't disable the test, skip it, or mark it as expected failure.

### Phase reference

See `docs/specs/implementation-plan.md` for the full phased plan. Summary:

| Phase | Module | Branch | Has UI? |
|-------|--------|--------|---------|
| 0 | Project bootstrap | `phase/bootstrap` | Yes (scaffold) |
| 1 | Data layer | `phase/data-layer` | No |
| 2 | Container management | `phase/containers` | No |
| 3 | Orchestration + token optimization | `phase/orchestration` | No |
| 4 | Project lifecycle API + PO API | `phase/project-api` | No |
| 5 | React UI + token dashboard + PO chat | `phase/ui` | Yes |
| 6 | Integration + PO workspace | `phase/integration` | Yes |
| 7 | Local mode (optional) | `phase/local-mode` | No |

## Code Patterns

### Go patterns

- **Feature-folder structure**: each package in `internal/` is self-contained with its own types, store, service, and tests.
- **Constructor injection**: `NewFooService(db *sqlx.DB, cache *cache.Client, broker *broker.Client) *FooService`
- **Small interfaces**: define at the consumer, not the provider. One or two methods max.
- **Error wrapping**: `fmt.Errorf("creating project: %w", err)` — always add context.
- **Table-driven tests**: every test file uses this pattern.
- **Real database tests**: test against Postgres, not mocks. Docker Compose provides the test DB.

### React patterns

- **Vite + TypeScript + Tailwind + shadcn/ui + TanStack Query**
- **Feature folders**: `web/src/features/projects/`, `web/src/features/agents/`
- **API client**: auto-generated from OpenAPI spec via `@hey-api/openapi-ts`
- **WebSocket hooks**: custom hooks for real-time event streaming
- **Playwright E2E**: tests in `web/e2e/`, mandatory for all UI features

## Key Files

<!-- Updated as the project grows -->
- `docs/specs/design-specification.md` — Full architecture and data model
- `docs/specs/implementation-plan.md` — Phased build plan with all tasks
- `docs/specs/po-prompt-architecture.md` — PO prompt design, playbooks, session types
- `docs/specs/dev-agent-architecture.md` — Dev agent prompts, communication protocol
- `docs/specs/agent-cli-provider-reference.md` — Claude/Gemini/Codex CLI integration details

## Agents

This project includes specialized agents in `.claude/agents/`. Use the Agent tool with `subagent_type` to invoke them for focused work.

| Agent | When to use |
|-------|-------------|
| `code-reviewer` | Comprehensive code reviews, security vulnerabilities, best practices |
| `debugger` | Diagnose and fix bugs, root cause analysis, error log analysis |
| `refactoring-specialist` | Transform poorly structured code while preserving behavior |
| `qa-expert` | Quality assurance strategy, test planning, quality metrics |
| `test-automator` | Build automated test frameworks, CI/CD integration |
| `golang-pro` | Go concurrency, high-performance systems, idiomatic patterns |
| `typescript-pro` | Advanced TypeScript type system, full-stack type safety |
| `javascript-pro` | Modern JS, ES2023+, async patterns, performance |
| `backend-developer` | Server-side APIs, microservices, production architecture |
| `frontend-developer` | Complete frontend applications, multi-framework expertise |
| `react-specialist` | React 18+ optimization, state management, architecture |
| `postgres-pro` | PostgreSQL optimization, replication, enterprise features |

Agents are invoked automatically by Claude when a task matches their specialty, or manually via the Agent tool.

## Lessons Learned

This project maintains a living document at `.claude/lessons_learned.md`. See `docs/design-principles.md` § Git & Workflow for the full protocol.

## Environment Variables
See `.env.example` for all vars.

## Project Documentation
- Design principles: `docs/design-principles.md`
- Architecture decisions: `docs/decisions/`
- Design specs: `docs/specs/`
