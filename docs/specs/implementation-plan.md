# Foundry — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a spec-driven AI development platform that orchestrates teams of Claude Code agents in Docker containers, managed by a Go control plane and coordinated by a Product Owner agent.

**Architecture:** Hybrid orchestration — Go handles infrastructure (containers, task state, DAG resolution, health checks), Claude Code PO agent handles judgment calls (planning, task decomposition, review). RabbitMQ for messaging, Postgres for durable state, Redis for cache, React for the UI.

**Tech Stack:** Go, React, Postgres, Redis, RabbitMQ, Docker SDK for Go, WebSocket, Claude Code CLI

---

## How to use this plan

1. Read this document before starting any work.
2. Phases are sequential — complete each phase before starting the next.
3. Within a phase, tasks can sometimes be parallelized (noted where possible).
4. Each task lists exact files to create/modify, tests to write, and commands to run.
5. Update task status as you go. Commit after each task.

## Status legend

| Status | Meaning |
|--------|---------|
| `pending` | Not started, waiting for dependencies |
| `in progress` | Being actively worked on |
| `complete` | Done and committed |

---

## Phase 0: Project bootstrap

> Sequential. Must be completed before any module work begins.
> Branch: `phase/bootstrap`

### Task 0.1: Project scaffold and Go module

**Files:**
- Create: `go.mod`
- Create: `cmd/foundry/main.go`
- Create: `Makefile`
- Create: `.env.example`
- Create: `.gitignore`
- Create: `CLAUDE.md`

- [ ] **Step 1: Initialize Go module**

```bash
mkdir -p ~/Documents/claude_workspace/foundry
cd ~/Documents/claude_workspace/foundry
go mod init github.com/cenron/foundry
```

- [ ] **Step 2: Create project directory structure**

```
foundry/
├── cmd/foundry/            # Entry point
├── internal/
│   ├── config/             # Configuration loading
│   ├── database/           # Postgres connection, migrations
│   ├── cache/              # Redis client
│   ├── broker/             # RabbitMQ client
│   ├── container/          # Docker container management
│   ├── orchestrator/       # DAG resolver, task state machine
│   ├── agent/              # Agent lifecycle, library loader
│   ├── project/            # Project CRUD, spec management
│   ├── event/              # Event logging, routing
│   ├── api/                # HTTP handlers, WebSocket
│   └── shared/             # Shared types, helpers
├── web/                    # React frontend
├── build/                  # Dockerfile, docker-compose
├── migrations/             # SQL migration files
├── docs/                   # Specs, plans
└── scripts/                # Dev scripts
```

```bash
mkdir -p cmd/foundry internal/{config,database,cache,broker,container,orchestrator,agent,project,event,api,shared} web build migrations docs scripts
```

- [ ] **Step 3: Create entry point stub**

```go
// cmd/foundry/main.go
package main

import "fmt"

func main() {
    fmt.Println("foundry starting...")
}
```

- [ ] **Step 4: Create Makefile**

Targets: `build`, `run`, `test`, `lint`, `docker-up`, `docker-down`, `migrate-up`, `migrate-down`

- [ ] **Step 5: Create .env.example**

```env
DATABASE_URL=postgres://foundry:foundry@localhost:5432/foundry?sslmode=disable
REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
ANTHROPIC_API_KEY=
GIT_TOKEN=
FOUNDRY_AGENT_LIBRARY=./agents
FOUNDRY_SSH_KEY_PATH=~/.ssh
```

- [ ] **Step 6: Create project CLAUDE.md**

Project-specific conventions, build commands, module boundaries.

- [ ] **Step 7: Create .gitignore, initialize git**

```bash
git init
git add -A
git commit -m "feat: scaffold Foundry project structure"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.2: Docker Compose dev infrastructure

**Files:**
- Create: `build/docker-compose.yml`
- Create: `build/Dockerfile`

- [ ] **Step 1: Write docker-compose.yml**

Services: Postgres 16, Redis 7, RabbitMQ 3 (with management plugin). Volumes for data persistence. Expose ports for local dev.

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: foundry
      POSTGRES_PASSWORD: foundry
      POSTGRES_DB: foundry
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest

volumes:
  pgdata:
```

- [ ] **Step 2: Write Dockerfile for Foundry backend**

Multi-stage build: Go builder stage, minimal runtime with the binary.

- [ ] **Step 3: Verify infrastructure starts**

```bash
make docker-up
# Verify: postgres, redis, rabbitmq all healthy
make docker-down
```

- [ ] **Step 4: Commit**

```bash
git add build/
git commit -m "feat: add Docker Compose dev infrastructure"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.3: Configuration module

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing test for config loading**

Test that config loads from environment variables with defaults.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/config/ -v
```

- [ ] **Step 3: Implement config struct and loader**

Single config struct with all settings: database URL, Redis URL, RabbitMQ URL, API port, Anthropic API key, git token, agent library path, SSH key path. Load from env vars with sensible defaults.

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/config/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add configuration module with env var loading"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.4: Database package

**Files:**
- Create: `internal/database/database.go`
- Create: `internal/database/migrate.go`
- Create: `internal/database/database_test.go`

- [ ] **Step 1: Write failing test for DB connection**

Test that a pool connects to Postgres and can ping.

- [ ] **Step 2: Implement database connection pool**

Use `jmoiron/sqlx` with the `lib/pq` driver. Accept config, return `*sqlx.DB`. Context-aware with timeout.

- [ ] **Step 3: Write failing test for migration runner**

Test that migrations run in order and are idempotent.

- [ ] **Step 4: Implement migration runner**

Read `.sql` files from `migrations/` directory, execute in order, track applied migrations in a `schema_migrations` table.

- [ ] **Step 5: Run all tests**

```bash
make docker-up
go test ./internal/database/ -v
```

- [ ] **Step 6: Commit**

```bash
git add internal/database/ migrations/
git commit -m "feat: add database package with connection pool and migration runner"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.5: Redis cache package

**Files:**
- Create: `internal/cache/cache.go`
- Create: `internal/cache/cache_test.go`

- [ ] **Step 1: Write failing test for cache get/set**

- [ ] **Step 2: Implement Redis client wrapper**

Thin wrapper around `go-redis/v9`. Methods: `Get`, `Set` (with TTL), `Delete`, `Exists`. JSON serialization for complex values.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/cache/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/cache/
git commit -m "feat: add Redis cache package"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.6: RabbitMQ broker package

**Files:**
- Create: `internal/broker/broker.go`
- Create: `internal/broker/exchanges.go`
- Create: `internal/broker/broker_test.go`

- [ ] **Step 1: Write failing test for publish/subscribe**

Test that a message published to an exchange is received by a subscriber.

- [ ] **Step 2: Implement broker client**

Connect to RabbitMQ, declare exchanges, publish messages, subscribe to queues. All exchanges are topic exchanges with project-scoped routing keys.

Define exchange constants:
- `foundry.commands` — topic exchange for agent commands. Routing key: `commands.<project_id>.<agent_id>`. Sidecar subscribes to `commands.<project_id>.*`.
- `foundry.events` — topic exchange for agent events. Routing key: `events.<project_id>.<event_type>`. Control plane subscribes to `events.<project_id>.*`.
- `foundry.logs` — topic exchange for agent log streaming. Routing key: `logs.<project_id>.<agent_id>`.

- [ ] **Step 3: Write test for topic routing**

Test that messages published to `logs.agent-123` are received by a subscriber on `logs.*`.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/broker/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/broker/
git commit -m "feat: add RabbitMQ broker package with topic and direct exchanges"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.7: Shared types

**Files:**
- Create: `internal/shared/types.go`
- Create: `internal/shared/errors.go`

- [ ] **Step 1: Define shared types**

ID type (UUID), timestamp helpers, status enums for Project, Task, Agent, Spec approval. Pagination params and response wrapper.

- [ ] **Step 2: Define error types**

`NotFoundError`, `ValidationError`, `ConflictError`. Each implements `error` interface and carries context.

- [ ] **Step 3: Commit**

```bash
git add internal/shared/
git commit -m "feat: add shared types and error definitions"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.8: API server skeleton

**Files:**
- Create: `internal/api/server.go`
- Create: `internal/api/middleware.go`
- Create: `internal/api/response.go`
- Create: `internal/api/websocket.go`
- Create: `internal/api/server_test.go`

- [ ] **Step 1: Write failing test for health endpoint**

Test that `GET /api/health` returns 200 with status "ok".

- [ ] **Step 2: Implement API server**

Use `chi` router. Set up middleware (logging, recovery, CORS). Health endpoint. JSON response helpers (`RespondJSON`, `RespondError`). Server struct that accepts dependencies via constructor.

- [ ] **Step 3: Write WebSocket hub stub**

WebSocket connection manager that will broadcast events to connected UI clients. Hub struct with `Register`, `Unregister`, `Broadcast` methods. Use `gorilla/websocket`.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/api/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/api/
git commit -m "feat: add API server skeleton with health endpoint and WebSocket hub"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.9: Composition root

**Files:**
- Modify: `cmd/foundry/main.go`

- [ ] **Step 1: Wire up the composition root**

Load config, connect to Postgres, Redis, RabbitMQ. Initialize API server. Start HTTP listener. Graceful shutdown on SIGINT/SIGTERM.

- [ ] **Step 2: Verify the server starts**

```bash
make docker-up
make run
# Verify: http://localhost:8080/api/health returns 200
```

- [ ] **Step 3: Commit**

```bash
git add cmd/foundry/main.go
git commit -m "feat: wire up composition root with graceful shutdown"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

### Task 0.10: Frontend scaffold

**Files:**
- Create: `web/package.json`
- Create: `web/vite.config.ts`
- Create: `web/tsconfig.json`
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/api/client.ts`
- Create: `web/src/api/websocket.ts`
- Create: `web/src/components/Layout.tsx`

- [ ] **Step 1: Scaffold React + Vite + TypeScript project**

```bash
cd web
npm create vite@latest . -- --template react-ts
npm install
```

- [ ] **Step 2: Install dependencies**

```bash
npm install react-router-dom @tanstack/react-query tailwindcss @tailwindcss/vite
```

- [ ] **Step 3: Configure Tailwind, TanStack Query, React Router**

Set up providers in `main.tsx`. Create basic route structure with placeholder pages: `/` (project list), `/projects/:id` (dashboard), `/projects/:id/agents/:agentId` (agent detail).

- [ ] **Step 4: Create API client module**

Fetch wrapper with base URL config. WebSocket connection manager with reconnect logic.

- [ ] **Step 5: Create Layout component**

App shell with nav bar (Foundry logo, minimal nav). Main content area.

- [ ] **Step 6: Configure Vite proxy for dev**

Proxy `/api` and `/ws` to Go backend on port 8080.

- [ ] **Step 7: Verify frontend starts**

```bash
npm run dev
# Verify: http://localhost:5173 loads
```

- [ ] **Step 8: Commit**

```bash
git add web/
git commit -m "feat: scaffold React frontend with routing and API client"
```

**Status:** `pending`
**Branch:** `phase/bootstrap`

---

## Phase 1: Data layer

> Depends on: Phase 0 complete
> Branch: `phase/data-layer`

### Task 1.1: Database migrations

**Files:**
- Create: `migrations/001_create_projects.sql`
- Create: `migrations/002_create_specs.sql`
- Create: `migrations/003_create_tasks.sql`
- Create: `migrations/004_create_agents.sql`
- Create: `migrations/005_create_events.sql`
- Create: `migrations/006_create_artifacts.sql`

- [ ] **Step 1: Write migration for projects table**

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    repo_url TEXT NOT NULL DEFAULT '',
    team_composition JSONB NOT NULL DEFAULT '[]',
    container_id TEXT,
    risk_profile_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- [ ] **Step 2: Write migration for specs table**

Foreign key to projects. Content as TEXT, token_estimate as INTEGER, agent_count as INTEGER, approval_status as TEXT.

- [ ] **Step 3: Write migration for tasks table**

Foreign key to projects and specs. depends_on as UUID array. Include `risk_level TEXT NOT NULL DEFAULT 'medium'`, `model_tier TEXT NOT NULL DEFAULT 'sonnet'`, `token_usage INTEGER NOT NULL DEFAULT 0`, `automation_eligible BOOLEAN NOT NULL DEFAULT false`. Index on project_id and status.

- [ ] **Step 4: Write migration for agents table**

Foreign key to projects. Include `worktree_path TEXT`, `branch_name TEXT`, `process_id TEXT`. Index on project_id and status.

- [ ] **Step 5: Write migration for risk_profiles table**

```sql
CREATE TABLE risk_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    name TEXT NOT NULL,
    low_criteria JSONB NOT NULL DEFAULT '{}',
    medium_criteria JSONB NOT NULL DEFAULT '{}',
    high_criteria JSONB NOT NULL DEFAULT '{}',
    model_routing JSONB NOT NULL DEFAULT '{"claude": {"low": "haiku", "medium": "sonnet", "high": "opus"}}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Insert a default global risk profile (project_id NULL).

- [ ] **Step 6: Write migration for events table**

Foreign keys to projects, tasks (nullable), agents (nullable). Payload as JSONB. Index on project_id and created_at. This is an append-only table.

- [ ] **Step 7: Write migration for artifacts table**

Foreign keys to projects, tasks, agents. Index on project_id.

- [ ] **Step 8: Write migrations for spec_mutations, agent_messages, and po_sessions tables**

`spec_mutations`: foreign key to specs, `field_changed TEXT`, `reason TEXT`, `diff TEXT`, `created_at`. Append-only.

`agent_messages`: foreign key to project, `from_agent_id`, `to_agent_id`, `content TEXT`, `po_intervention TEXT`, `created_at`.

`po_sessions`: `id UUID`, foreign key to projects, `pid INTEGER`, `status TEXT` (starting, active, paused, stopped), `session_type TEXT` (planning, estimation, review, etc.), `workspace_path TEXT`, `started_at TIMESTAMPTZ`, `stopped_at TIMESTAMPTZ`. Tracks PO process lifecycle — used by the control plane to serialize system-triggered sessions per project.

- [ ] **Step 9: Run migrations**

```bash
make migrate-up
```

- [ ] **Step 10: Commit**

```bash
git add migrations/
git commit -m "feat: add database migrations for all entities"
```

**Status:** `pending`
**Branch:** `phase/data-layer`

---

### Task 1.2: Project repository

**Files:**
- Create: `internal/project/store.go`
- Create: `internal/project/models.go`
- Create: `internal/project/store_test.go`

- [ ] **Step 1: Define Project and Spec model structs**

Go structs with scan/serialize methods for database rows.

- [ ] **Step 2: Write failing tests for CRUD operations**

Create, GetByID, List, UpdateStatus for projects. Create, GetByProjectID, UpdateApproval for specs.

- [ ] **Step 3: Implement ProjectStore**

Hand-written SQL with sqlx. All queries use parameterized statements. Use `sqlx.Get`/`sqlx.Select` for struct scanning, named parameters where readable.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/project/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/project/
git commit -m "feat: add project and spec repository with CRUD operations"
```

**Status:** `pending`
**Branch:** `phase/data-layer`

---

### Task 1.3: Task repository

**Files:**
- Create: `internal/orchestrator/store.go`
- Create: `internal/orchestrator/models.go`
- Create: `internal/orchestrator/store_test.go`

- [ ] **Step 1: Define Task model struct**

Include depends_on as `[]uuid.UUID`, context_summary as nullable text.

- [ ] **Step 2: Write failing tests**

Create, GetByID, ListByProject, UpdateStatus, UpdateAssignment, GetUnblockedTasks (tasks where all dependencies are status `done`).

- [ ] **Step 3: Implement TaskStore**

The `GetUnblockedTasks` query is the key one — it joins the task's dependencies and filters to tasks where all deps have status `done` and the task itself is `pending`.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/orchestrator/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/orchestrator/
git commit -m "feat: add task repository with DAG-aware unblocked query"
```

**Status:** `pending`
**Branch:** `phase/data-layer`

---

### Task 1.4: Agent and event repositories

**Files:**
- Create: `internal/agent/store.go`
- Create: `internal/agent/models.go`
- Create: `internal/agent/store_test.go`
- Create: `internal/event/store.go`
- Create: `internal/event/models.go`
- Create: `internal/event/store_test.go`

- [ ] **Step 1: Define Agent model struct**

- [ ] **Step 2: Write failing tests for agent CRUD**

Create, GetByID, ListByProject, UpdateStatus, UpdateHealth, UpdateCurrentTask.

- [ ] **Step 3: Implement AgentStore**

- [ ] **Step 4: Define Event and Artifact model structs**

- [ ] **Step 5: Write failing tests for event logging**

Create event, ListByProject (paginated, newest first), ListByAgent.

- [ ] **Step 6: Implement EventStore**

Append-only insert. Paginated list queries.

- [ ] **Step 7: Define Artifact model struct and write failing tests**

Create artifact, ListByProject, ListByTask, ListByAgent.

- [ ] **Step 8: Implement ArtifactStore**

Create (records path on shared volume, type, description). List queries with filtering by project, task, or agent.

- [ ] **Step 9: Run all tests**

```bash
go test ./internal/agent/ ./internal/event/ -v
```

- [ ] **Step 10: Commit**

```bash
git add internal/agent/ internal/event/
git commit -m "feat: add agent, event, and artifact repositories"
```

**Status:** `pending`
**Branch:** `phase/data-layer`

---

## Phase 2: Container management

> Depends on: Phase 1 complete
> Branch: `phase/containers`

### Task 2.1: Docker team container manager

**Files:**
- Create: `internal/container/manager.go`
- Create: `internal/container/config.go`
- Create: `internal/container/manager_test.go`

- [ ] **Step 1: Write failing test for team container creation**

Test that creating a team container produces the right Docker API call with correct volume mounts (workspace, shared, agents library, SSH keys).

- [ ] **Step 2: Implement ContainerManager**

Uses Docker SDK for Go (`github.com/docker/docker/client`). One container per project team — all agents run as processes inside it, using git worktrees for isolation. Methods:

- `CreateTeam(ctx, TeamContainerConfig) (containerID, error)` — pulls team image, creates container with env vars, volume mounts, network
- `StartTeam(ctx, containerID) error`
- `StopTeam(ctx, containerID) error` — sends SIGTERM, waits for graceful shutdown
- `RemoveTeam(ctx, containerID) error`
- `GetStatus(ctx, containerID) (ContainerStatus, error)`
- `StreamLogs(ctx, containerID) (io.ReadCloser, error)`
- `ExecInTeam(ctx, containerID, cmd []string) (output string, error)` — execute commands inside the running team container (for worktree setup, agent launches)

- [ ] **Step 3: Define TeamContainerConfig**

```go
type TeamContainerConfig struct {
    ProjectID      string
    RepoURL        string
    AnthropicKey   string
    RabbitMQURL    string
    GitToken       string
    NetworkName    string
    SharedVolPath  string   // host path for shared volume (~/foundry/projects/<name>/shared/)
    AgentLibPath   string   // host path for agent library
    SSHKeyPath     string   // host path for SSH keys
    TeamComposition []string // agent roles to set up
}
```

- [ ] **Step 4: Write test for volume mount configuration**

Verify that workspace, shared, agents library, and SSH key directories are mounted correctly. Shared volume must be bind-mounted to host path so PO can access it.

- [ ] **Step 5: Run tests**

```bash
go test ./internal/container/ -v
```

- [ ] **Step 6: Commit**

```bash
git add internal/container/
git commit -m "feat: add Docker team container manager with one-container-per-project model"
```

**Status:** `pending`
**Branch:** `phase/containers`

---

### Task 2.2: Team container image

**Files:**
- Create: `build/agent/Dockerfile`
- Create: `build/agent/entrypoint.sh`
- Create: `build/agent/supervisor.sh`
- Create: `build/agent/sidecar.js`
- Create: `build/agent/heartbeat.sh`
- Create: `build/agent/workspace/CLAUDE.md`
- Create: `build/agent/workspace/.claude/languages/go.md`
- Create: `build/agent/workspace/.claude/languages/node.md`
- Create: `build/agent/workspace/.claude/languages/python.md`
- Create: `build/agent/workspace/.claude/frameworks/react.md`

- [ ] **Step 1: Create the Foundry agent workspace template**

Build a custom workspace environment for orchestrated agents. Use `github.com/cenron/claude-env` as reference for coding conventions, but tailor for Foundry agent context. The CLAUDE.md should include:
- Coding standards, verification loops, design patterns (adapted from claude-env)
- Agent-specific instructions: how to check for pause signals at `/foundry/state/pause-signal`, how to write artifacts to `/shared`, how to report progress, how to write context summaries on pause
- Language/framework convention files selected per agent role
- Task-specific instructions section that the entrypoint populates from the PO's assignment

- [ ] **Step 2: Write Dockerfile for team container image**

Ubuntu base. Install Git, Node.js, common dev tools. Install Claude Code CLI globally (`npm install -g @anthropic-ai/claude-code`). Install Go (for backend agents). Copy entrypoint, supervisor, sidecar, and heartbeat scripts. Bake in the workspace template at `/foundry/workspace-template/`.

- [ ] **Step 3: Write entrypoint script**

The entrypoint bootstraps the team container:
1. Clones the repo into `/workspace`
2. Starts `sidecar.js` (RabbitMQ client) in background
3. Starts `heartbeat.sh` in background
4. Reads team composition from `/shared/team.json`
5. For each agent role in the team:
   a. Creates git worktree at `/worktrees/<role>/` on the agent's branch
   b. Copies role-specific workspace (language/framework files, agent-role.md) into the worktree
   c. Generates composite CLAUDE.md (base + project overlay + role)
   d. Launches agent CLI process in the worktree directory
   e. Registers process with supervisor
6. Supervisor monitors all agent processes, reports status via sidecar

- [ ] **Step 4: Write process supervisor**

`supervisor.sh` manages all agent processes inside the team container:
- Launch and register new agent processes (including on-demand QA agents)
- Monitor process health (alive, responsive)
- Handle clean pause: relay pause signals, wait for agent to commit and write context summary, then stop the process
- On unexpected crash: log the failure, notify the sidecar, do not auto-restart (PO decides)

- [ ] **Step 5: Write RabbitMQ sidecar**

`sidecar.js` (Node.js) runs alongside agents in the container:
- Connects to RabbitMQ using `RABBITMQ_URL`
- Subscribes to `foundry.commands` queue filtered by project ID
- Routes commands to the appropriate agent process via supervisor
- Publishes events to `foundry.events` exchange
- Streams agent stdout to `foundry.logs` exchange with routing key `logs.<project_id>.<agent_id>`
- Relays agent-to-agent messages (PO receives copies of all messages)

- [ ] **Step 6: Build and test the team image locally**

```bash
docker build -t foundry-team:dev -f build/agent/Dockerfile build/agent/
docker run --rm foundry-team:dev claude --version
```

- [ ] **Step 7: Commit**

```bash
git add build/agent/
git commit -m "feat: add team container image with supervisor, sidecar, and workspace template"
```

**Status:** `pending`
**Branch:** `phase/containers`

---

### Task 2.3: Health monitor

**Files:**
- Create: `internal/container/health.go`
- Create: `internal/container/health_test.go`

- [ ] **Step 1: Write failing test for health check logic**

Test that 3 missed heartbeats (30s) marks agent as unhealthy.

- [ ] **Step 2: Implement HealthMonitor**

Runs as a goroutine. Every 10 seconds, checks the team container's heartbeat. The sidecar inside the container monitors individual agent processes and reports their health. If 3 consecutive heartbeat misses from the container, marks all agents as `unhealthy` in the agent store and publishes events to RabbitMQ.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/container/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/container/health.go internal/container/health_test.go
git commit -m "feat: add health monitor with heartbeat-based detection"
```

**Status:** `pending`
**Branch:** `phase/containers`

---

### Task 2.4: Workspace environment builder

**Files:**
- Create: `internal/agent/workspace.go`
- Create: `internal/agent/workspace_test.go`

- [ ] **Step 1: Write failing test for workspace assembly**

Test that given a role and project config, the builder produces the right set of files (CLAUDE.md, language files, framework files, agent definition).

- [ ] **Step 2: Implement WorkspaceBuilder**

Reads the Foundry workspace template from the agent image (baked in at `/foundry/workspace-template/`). Assembles per-agent:
- Base CLAUDE.md from the template
- `.claude/languages/` — filtered by role (Go agent gets `go.md`, React agent gets `node.md`)
- `.claude/frameworks/` — filtered by role
- `.claude/lessons_learned.md` — starts empty per project, accumulates during execution
- Agent role definition from the agent library (copied to `.claude/agent-role.md`)

Generates a project-specific CLAUDE.md overlay with:
- Project name, description, and repo context
- Module boundaries and conventions from the spec
- Task assignment details from the PO

Merges the base CLAUDE.md + overlay into a composite CLAUDE.md. Writes all files to the agent's `/workspace` volume after the repo is cloned.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/agent/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/agent/workspace.go internal/agent/workspace_test.go
git commit -m "feat: add workspace environment builder for agent worktrees"
```

**Status:** `pending`
**Branch:** `phase/containers`

---

### Task 2.5: Agent library loader

**Files:**
- Create: `internal/agent/library.go`
- Create: `internal/agent/library_test.go`

- [ ] **Step 1: Write failing test for loading agent definitions**

Test that markdown files with YAML frontmatter are parsed into AgentDefinition structs.

- [ ] **Step 2: Implement AgentLibrary**

Scans a directory for `.md` files. Parses YAML frontmatter to extract: name, description, tools, model. Stores the full file content for injection into containers.

Methods:
- `LoadAll() ([]AgentDefinition, error)`
- `GetByName(name string) (AgentDefinition, error)`
- `ListRoles() []string`

- [ ] **Step 3: Run tests**

```bash
go test ./internal/agent/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/agent/library.go internal/agent/library_test.go
git commit -m "feat: add agent library loader for markdown definitions"
```

**Status:** `pending`
**Branch:** `phase/containers`

---

## Phase 3: Orchestration engine

> Depends on: Phase 2 complete
> Branch: `phase/orchestration`

### Task 3.1: Task state machine

**Files:**
- Create: `internal/orchestrator/statemachine.go`
- Create: `internal/orchestrator/statemachine_test.go`

- [ ] **Step 1: Write failing tests for valid state transitions**

Test that: pending → assigned, assigned → in_progress, in_progress → paused, in_progress → review, review → done, paused → assigned. Test that invalid transitions are rejected.

- [ ] **Step 2: Implement TaskStateMachine**

Validates state transitions. On each transition, persists to DB and publishes an event to RabbitMQ.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/orchestrator/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/orchestrator/statemachine.go internal/orchestrator/statemachine_test.go
git commit -m "feat: add task state machine with transition validation"
```

**Status:** `pending`
**Branch:** `phase/orchestration`

---

### Task 3.2: DAG resolver

**Files:**
- Create: `internal/orchestrator/dag.go`
- Create: `internal/orchestrator/dag_test.go`

- [ ] **Step 1: Write failing tests for dependency resolution**

Test scenarios:
- Task with no dependencies is immediately unblocked
- Task with one dependency is blocked until dep is done
- Task with multiple dependencies is blocked until all are done
- Circular dependency detection

- [ ] **Step 2: Implement DAGResolver**

Methods:
- `GetUnblockedTasks(ctx, projectID) ([]Task, error)` — delegates to TaskStore query
- `ValidateDependencies(tasks []Task) error` — checks for cycles using topological sort
- `GetCriticalPath(tasks []Task) []Task` — identifies the longest chain (useful for PO estimation)

- [ ] **Step 3: Run tests**

```bash
go test ./internal/orchestrator/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/orchestrator/dag.go internal/orchestrator/dag_test.go
git commit -m "feat: add DAG resolver with cycle detection and critical path"
```

**Status:** `pending`
**Branch:** `phase/orchestration`

---

### Task 3.3: Orchestrator service

**Files:**
- Create: `internal/orchestrator/service.go`
- Create: `internal/orchestrator/service_test.go`

- [ ] **Step 1: Write failing tests for the orchestration loop**

Test that when a task completes, the orchestrator checks for newly unblocked tasks and assigns them to available agents.

- [ ] **Step 2: Implement OrchestratorService**

The main orchestration loop. Subscribes to RabbitMQ events. On `task_completed`:
1. Update task status to `done`
2. Query for newly unblocked tasks
3. Find available agents (status `active`, no current task)
4. Match tasks to agents by role
5. Assign tasks via the task state machine
6. Send assignment commands to agents via RabbitMQ

On `agent_unhealthy`:
1. Mark the agent's current task as `paused`
2. Publish event for PO to decide next action

On `pause_agent` command:
1. Send pause command to agent via RabbitMQ
2. Wait for agent to report clean stop
3. Update agent status to `paused`

On `resume_agent` command:
1. Restart the agent process inside the team container (or restart the team container if it was stopped)
2. Reassign the paused task

- [ ] **Step 3: Run tests**

```bash
go test ./internal/orchestrator/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/orchestrator/service.go internal/orchestrator/service_test.go
git commit -m "feat: add orchestrator service with event-driven task assignment"
```

**Status:** `pending`
**Branch:** `phase/orchestration`

---

### Task 3.4: Event router

**Files:**
- Create: `internal/event/router.go`
- Create: `internal/event/router_test.go`

- [ ] **Step 1: Write failing test for event routing**

Test that events from RabbitMQ are persisted to Postgres and forwarded to WebSocket clients.

- [ ] **Step 2: Implement EventRouter**

Subscribes to `foundry.events` exchange. For each event:
1. Persist to Postgres via EventStore
2. Update Redis cache (agent state, task state)
3. Forward to WebSocket hub for UI broadcast

Also subscribes to `foundry.logs` exchange:
1. Forward log lines to WebSocket hub (scoped to agent ID)
2. Optionally batch-persist to Postgres for history

- [ ] **Step 3: Run tests**

```bash
go test ./internal/event/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/event/router.go internal/event/router_test.go
git commit -m "feat: add event router bridging RabbitMQ to WebSocket and Postgres"
```

**Status:** `pending`
**Branch:** `phase/orchestration`

---

### Task 3.5: Model tier resolution and token tracking

**Files:**
- Create: `internal/agent/tier.go`
- Create: `internal/agent/tier_test.go`
- Create: `internal/agent/pricing.go`

- [ ] **Step 1: Write failing tests for tier resolution**

Test the full chain: task risk_level → risk_profile model_routing → abstract tier → provider model name. Test the `min_model` floor check (e.g., security-reviewer should never get downgraded below sonnet even on low-risk tasks).

- [ ] **Step 2: Implement TierResolver**

```go
type TierResolver struct {
    riskProfile RiskProfile
}

// Resolve returns the concrete model string for a given task risk level and agent role
func (r *TierResolver) Resolve(riskLevel string, role AgentDefinition) string
```

Reads `model_routing` from the risk profile, checks the role's `min_model` floor, returns the resolved tier.

- [ ] **Step 3: Implement price table and cost calculator**

Configurable price table (loaded from config, not hardcoded). Calculate cost from token usage:

```go
type ModelPricing struct {
    InputPer1K  float64
    OutputPer1K float64
    CachedPer1K float64
}

func CalculateCost(usage TokenUsage, model string, table PriceTable) float64
```

For Claude: use the `total_cost_usd` from the result event directly. Price table is the fallback for providers that don't report cost.

- [ ] **Step 4: Implement token budget tracker**

Tracks per-task and per-project token consumption. Publishes budget utilization events at 50%, 75%, and 90% thresholds for the PO to act on.

```go
type BudgetTracker struct {
    store     TokenUsageStore
    publisher EventPublisher
}

func (t *BudgetTracker) RecordUsage(ctx context.Context, taskID, projectID string, usage TokenUsage) error
func (t *BudgetTracker) GetProjectUtilization(ctx context.Context, projectID string) (BudgetStatus, error)
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/agent/ -v -run TestTier
```

- [ ] **Step 6: Commit**

```bash
git add internal/agent/tier.go internal/agent/tier_test.go internal/agent/pricing.go
git commit -m "feat: add model tier resolution, price table, and token budget tracking"
```

**Status:** `pending`
**Branch:** `phase/orchestration`

---

### Task 3.6: Response cache

**Files:**
- Create: `internal/cache/response.go`
- Create: `internal/cache/response_test.go`

- [ ] **Step 1: Write failing test for response caching**

Test that identical prompts (same hash) return cached responses. Test TTL expiry. Test per-project scoping.

- [ ] **Step 2: Implement ResponseCache**

Redis-backed cache keyed by `hash(project_id + prompt + model + temperature)`. Short TTL (5-10 minutes). Per-project scoped. Best-effort — misses fall through to the API. Track hit rates as event metrics.

```go
type ResponseCache struct {
    redis *cache.Client
}

func (c *ResponseCache) Get(ctx context.Context, key ResponseCacheKey) (string, bool, error)
func (c *ResponseCache) Set(ctx context.Context, key ResponseCacheKey, response string, ttl time.Duration) error
func (c *ResponseCache) HitRate(ctx context.Context, projectID string) (float64, error)
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/cache/ -v -run TestResponse
```

- [ ] **Step 4: Commit**

```bash
git add internal/cache/response.go internal/cache/response_test.go
git commit -m "feat: add per-project response cache with TTL and hit rate tracking"
```

**Status:** `pending`
**Branch:** `phase/orchestration`

---

## Phase 4: Project lifecycle API

> Depends on: Phase 3 complete
> Branch: `phase/project-api`

### Task 4.1: Project API handlers

**Files:**
- Create: `internal/api/projects.go`
- Create: `internal/api/projects_test.go`

- [ ] **Step 1: Write failing tests for project CRUD endpoints**

```
POST   /api/projects          — create project
GET    /api/projects          — list projects
GET    /api/projects/:id      — get project detail
PATCH  /api/projects/:id      — update project (name, description)
```

- [ ] **Step 2: Implement handlers**

Each handler validates input, calls the project store, returns JSON. Include project stats in the detail response (task counts by status, active agent count).

- [ ] **Step 3: Run tests**

```bash
go test ./internal/api/ -v -run TestProjects
```

- [ ] **Step 4: Commit**

```bash
git add internal/api/projects.go internal/api/projects_test.go
git commit -m "feat: add project CRUD API endpoints"
```

**Status:** `pending`
**Branch:** `phase/project-api`

---

### Task 4.2: Spec API handlers

**Files:**
- Create: `internal/api/specs.go`
- Create: `internal/api/specs_test.go`

- [ ] **Step 1: Write failing tests for spec endpoints**

```
GET    /api/projects/:id/spec       — get spec for project
PUT    /api/projects/:id/spec       — update spec content
POST   /api/projects/:id/spec/approve — approve spec (moves project to approved)
POST   /api/projects/:id/spec/reject  — reject spec (moves project back to planning)
```

- [ ] **Step 2: Implement handlers**

Approval endpoint validates that spec has content and a token estimate before allowing approval. Updates both spec approval_status and project status atomically.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/api/ -v -run TestSpecs
```

- [ ] **Step 4: Commit**

```bash
git add internal/api/specs.go internal/api/specs_test.go
git commit -m "feat: add spec API endpoints with approval gate"
```

**Status:** `pending`
**Branch:** `phase/project-api`

---

### Task 4.3: Task API handlers

**Files:**
- Create: `internal/api/tasks.go`
- Create: `internal/api/tasks_test.go`

- [ ] **Step 1: Write failing tests for task endpoints**

```
GET    /api/projects/:id/tasks          — list tasks (filterable by status)
GET    /api/projects/:id/tasks/:taskId  — get task detail
```

Tasks are created by the PO, not via API. The API is read-only for the UI kanban board.

- [ ] **Step 2: Implement handlers**

Include dependency information and assigned agent details in responses.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/api/ -v -run TestTasks
```

- [ ] **Step 4: Commit**

```bash
git add internal/api/tasks.go internal/api/tasks_test.go
git commit -m "feat: add task API endpoints for kanban board"
```

**Status:** `pending`
**Branch:** `phase/project-api`

---

### Task 4.4: Agent API handlers

**Files:**
- Create: `internal/api/agents.go`
- Create: `internal/api/agents_test.go`

- [ ] **Step 1: Write failing tests for agent endpoints**

```
GET    /api/projects/:id/agents            — list agents for project
GET    /api/projects/:id/agents/:agentId   — get agent detail
POST   /api/projects/:id/agents/:agentId/pause   — pause agent (clean)
POST   /api/projects/:id/agents/:agentId/resume  — resume agent
POST   /api/projects/:id/start             — start project (spin up team)
POST   /api/projects/:id/pause             — pause all agents
POST   /api/projects/:id/resume            — resume all agents
```

- [ ] **Step 2: Implement handlers**

Pause sends a command via RabbitMQ to the orchestrator. Start validates project is in `approved` status, then triggers the orchestrator to spin up the team. Agent detail includes current task, health status, and container info.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/api/ -v -run TestAgents
```

- [ ] **Step 4: Commit**

```bash
git add internal/api/agents.go internal/api/agents_test.go
git commit -m "feat: add agent API endpoints with pause/resume controls"
```

**Status:** `pending`
**Branch:** `phase/project-api`

---

### Task 4.5: Token usage and risk profile API handlers

**Files:**
- Create: `internal/api/usage.go`
- Create: `internal/api/usage_test.go`
- Create: `internal/api/riskprofile.go`
- Create: `internal/api/riskprofile_test.go`

- [ ] **Step 1: Write failing tests for token usage endpoints**

```
GET /api/projects/:id/usage — per-task token breakdown, budget utilization, model tier distribution, cost savings
```

- [ ] **Step 2: Implement usage handler**

Aggregates token_usage from tasks, calculates budget utilization against estimate, computes cost savings (actual vs. opus-only baseline).

- [ ] **Step 3: Write failing tests for risk profile endpoints**

```
GET /api/projects/:id/risk-profile — get risk profile (or global default)
PUT /api/projects/:id/risk-profile — update risk profile (model_routing, criteria)
```

- [ ] **Step 4: Implement risk profile handlers**

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/api/ -v -run "TestUsage|TestRiskProfile"
git add internal/api/usage.go internal/api/usage_test.go internal/api/riskprofile.go internal/api/riskprofile_test.go
git commit -m "feat: add token usage and risk profile API endpoints"
```

**Status:** `pending`
**Branch:** `phase/project-api`

---

### Task 4.6: PO chat API handler

**Files:**
- Create: `internal/api/po.go`
- Create: `internal/api/po_test.go`

- [ ] **Step 1: Write failing tests for PO chat endpoints**

```
POST /api/projects/:id/po/chat         — send message to PO, stream response
DELETE /api/projects/:id/po/chat        — close PO chat session
GET /api/projects/:id/po/status         — check if PO session is active
POST /api/projects/:id/po/planning      — start planning session
POST /api/projects/:id/po/estimation    — trigger estimation after spec approval
```

- [ ] **Step 2: Implement PO chat handler**

Accepts user message, delegates to POSessionManager to launch or continue a PO session. Streams PO response events to the client via WebSocket. Planning and estimation endpoints trigger the appropriate session types.

- [ ] **Step 3: Run tests and commit**

```bash
go test ./internal/api/ -v -run TestPO
git add internal/api/po.go internal/api/po_test.go
git commit -m "feat: add PO chat and session API endpoints"
```

**Status:** `pending`
**Branch:** `phase/project-api`

---

### Task 4.7: WebSocket event streaming

**Files:**
- Modify: `internal/api/websocket.go`
- Create: `internal/api/websocket_test.go`

- [ ] **Step 1: Write failing test for WebSocket event broadcast**

Test that when an event is published, connected WebSocket clients receive it.

- [ ] **Step 2: Implement WebSocket endpoints**

```
WS /ws/projects/:id/events  — stream project events (task moves, agent status)
WS /ws/agents/:agentId/logs — stream agent logs
```

Clients connect, receive JSON events. Support filtering by event type. The EventRouter pushes events into the WebSocket hub, which fans out to connected clients.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/api/ -v -run TestWebSocket
```

- [ ] **Step 4: Commit**

```bash
git add internal/api/websocket.go internal/api/websocket_test.go
git commit -m "feat: add WebSocket endpoints for real-time event and log streaming"
```

**Status:** `pending`
**Branch:** `phase/project-api`

---

### Task 4.8: Agent library API

**Files:**
- Create: `internal/api/library.go`
- Create: `internal/api/library_test.go`

- [ ] **Step 1: Write failing tests**

```
GET /api/agents/library — list available agent roles from the library
```

- [ ] **Step 2: Implement handler**

Returns the list of agent definitions (name, description, model) from the AgentLibrary. The UI uses this to display available roles and the PO uses it for team composition.

- [ ] **Step 3: Run tests and commit**

```bash
go test ./internal/api/ -v -run TestLibrary
git add internal/api/library.go internal/api/library_test.go
git commit -m "feat: add agent library API endpoint"
```

**Status:** `pending`
**Branch:** `phase/project-api`

---

## Phase 5: React UI

> Depends on: Phase 4 complete
> Branch: `phase/ui`

### Task 5.1: Project list page

**Files:**
- Create: `web/src/pages/ProjectList.tsx`
- Create: `web/src/components/ProjectCard.tsx`
- Create: `web/src/components/CreateProjectDialog.tsx`
- Create: `web/src/hooks/useProjects.ts`

- [ ] **Step 1: Create useProjects hook**

TanStack Query hook wrapping `GET /api/projects`. Auto-refetch on window focus.

- [ ] **Step 2: Create ProjectCard component**

Displays project name, status badge, task progress (X/Y done), active agent count. Primary action button based on status (Approve, View, Start).

- [ ] **Step 3: Create CreateProjectDialog**

Modal form with name, description, repo URL fields. Submits `POST /api/projects`.

- [ ] **Step 4: Create ProjectList page**

Grid of ProjectCards. "New Project" button opens the dialog. Empty state for no projects.

- [ ] **Step 5: Verify in browser**

```bash
cd web && npm run dev
# Navigate to http://localhost:5173, verify project list renders
```

- [ ] **Step 6: Commit**

```bash
git add web/src/pages/ProjectList.tsx web/src/components/ProjectCard.tsx web/src/components/CreateProjectDialog.tsx web/src/hooks/useProjects.ts
git commit -m "feat: add project list page with create dialog"
```

**Status:** `pending`
**Branch:** `phase/ui`

---

### Task 5.2: Kanban board component

**Files:**
- Create: `web/src/components/KanbanBoard.tsx`
- Create: `web/src/components/KanbanColumn.tsx`
- Create: `web/src/components/TaskCard.tsx`
- Create: `web/src/hooks/useTasks.ts`
- Create: `web/src/hooks/useProjectEvents.ts`

- [ ] **Step 1: Create useTasks hook**

TanStack Query hook wrapping `GET /api/projects/:id/tasks`.

- [ ] **Step 2: Create useProjectEvents hook**

WebSocket hook that connects to `/ws/projects/:id/events`. On task status change events, invalidates the tasks query to trigger re-fetch. Provides a stream of events for the activity feed.

- [ ] **Step 3: Create TaskCard component**

Displays task title, assigned agent (color-coded by role), branch name, blocked-by indicator. Border color matches agent role.

- [ ] **Step 4: Create KanbanColumn component**

Column header with status name and count. Renders a list of TaskCards. Columns: Backlog (pending + assigned), In Progress, Review, Done.

- [ ] **Step 5: Create KanbanBoard component**

Horizontal layout of KanbanColumns. Tasks auto-move between columns as WebSocket events arrive.

- [ ] **Step 6: Verify in browser**

- [ ] **Step 7: Commit**

```bash
git add web/src/components/Kanban*.tsx web/src/components/TaskCard.tsx web/src/hooks/useTasks.ts web/src/hooks/useProjectEvents.ts
git commit -m "feat: add real-time kanban board with WebSocket updates"
```

**Status:** `pending`
**Branch:** `phase/ui`

---

### Task 5.3: Agent status bar

**Files:**
- Create: `web/src/components/AgentStatusBar.tsx`
- Create: `web/src/components/AgentChip.tsx`
- Create: `web/src/hooks/useAgents.ts`

- [ ] **Step 1: Create useAgents hook**

TanStack Query hook wrapping `GET /api/projects/:id/agents`. Invalidates on agent status change events from WebSocket.

- [ ] **Step 2: Create AgentChip component**

Compact card showing: role name (color-coded), current task name, health indicator (green/yellow/red dot), pause button. Clicking the chip navigates to agent detail.

- [ ] **Step 3: Create AgentStatusBar component**

Horizontal scrollable bar of AgentChips. Shows above the kanban board on the project dashboard.

- [ ] **Step 4: Commit**

```bash
git add web/src/components/AgentStatusBar.tsx web/src/components/AgentChip.tsx web/src/hooks/useAgents.ts
git commit -m "feat: add agent status bar with health indicators"
```

**Status:** `pending`
**Branch:** `phase/ui`

---

### Task 5.4: Project dashboard page

**Files:**
- Create: `web/src/pages/ProjectDashboard.tsx`
- Create: `web/src/components/SpecViewer.tsx`
- Create: `web/src/hooks/useSpec.ts`

- [ ] **Step 1: Create useSpec hook**

TanStack Query hook for `GET /api/projects/:id/spec`.

- [ ] **Step 2: Create SpecViewer component**

Renders spec markdown content. Shows token estimate and approval status. Approve/Reject buttons when status is `estimated`.

- [ ] **Step 3: Create ProjectDashboard page**

Header: project name, status badge, "Pause All" / "Start" button, "View Spec" toggle.
Body: AgentStatusBar at top, KanbanBoard below.
Spec viewer slides in as a panel when toggled.

- [ ] **Step 4: Wire up pause/resume/start actions**

Pause All calls `POST /api/projects/:id/pause`. Start calls `POST /api/projects/:id/start`. Individual agent pause from AgentChip calls `POST /api/projects/:id/agents/:agentId/pause`.

- [ ] **Step 5: Verify in browser**

- [ ] **Step 6: Commit**

```bash
git add web/src/pages/ProjectDashboard.tsx web/src/components/SpecViewer.tsx web/src/hooks/useSpec.ts
git commit -m "feat: add project dashboard with kanban, agents, and spec viewer"
```

**Status:** `pending`
**Branch:** `phase/ui`

---

### Task 5.5: Agent detail page

**Files:**
- Create: `web/src/pages/AgentDetail.tsx`
- Create: `web/src/components/AgentInfo.tsx`
- Create: `web/src/components/LogStream.tsx`
- Create: `web/src/hooks/useAgentLogs.ts`

- [ ] **Step 1: Create useAgentLogs hook**

WebSocket hook connecting to `/ws/agents/:agentId/logs`. Maintains a buffer of the last 500 log lines. Auto-scrolls to bottom.

- [ ] **Step 2: Create AgentInfo component**

Left panel showing: role, current task, branch name, commit count, container status, health, list of artifacts produced.

- [ ] **Step 3: Create LogStream component**

Monospace terminal-style log viewer. Streams live output from the agent's Claude Code session. Color-coded by log type (info, success, error, command).

- [ ] **Step 4: Create AgentDetail page**

Split layout: AgentInfo on the left (250px), LogStream filling the rest. Header with agent role, back button, pause/resume control.

- [ ] **Step 5: Verify in browser**

- [ ] **Step 6: Commit**

```bash
git add web/src/pages/AgentDetail.tsx web/src/components/AgentInfo.tsx web/src/components/LogStream.tsx web/src/hooks/useAgentLogs.ts
git commit -m "feat: add agent detail page with live log streaming"
```

**Status:** `pending`
**Branch:** `phase/ui`

---

### Task 5.6: Token dashboard

**Files:**
- Create: `web/src/components/TokenDashboard.tsx`
- Create: `web/src/components/BudgetBar.tsx`
- Create: `web/src/components/ModelTierChart.tsx`
- Create: `web/src/hooks/useTokenUsage.ts`

- [ ] **Step 1: Create useTokenUsage hook**

TanStack Query hook wrapping `GET /api/projects/:id/usage`. Returns per-task token breakdown, total budget utilization, model tier distribution, and cost savings indicator.

- [ ] **Step 2: Create BudgetBar component**

Visual progress bar showing estimated vs. actual tokens consumed. Color changes at 75% (yellow) and 90% (red).

- [ ] **Step 3: Create ModelTierChart component**

Simple bar or pie chart showing what percentage of tasks ran on each model tier (haiku/sonnet/opus). Show cost savings: actual cost vs. what it would have cost on Opus-only.

- [ ] **Step 4: Create TokenDashboard component**

Combines BudgetBar, ModelTierChart, and a per-task token breakdown table. Integrated into the ProjectDashboard as a collapsible panel.

- [ ] **Step 5: Commit**

```bash
git add web/src/components/Token*.tsx web/src/components/BudgetBar.tsx web/src/components/ModelTierChart.tsx web/src/hooks/useTokenUsage.ts
git commit -m "feat: add token dashboard with budget utilization and cost savings"
```

**Status:** `pending`
**Branch:** `phase/ui`

---

### Task 5.7: Settings page

**Files:**
- Create: `web/src/pages/ProjectSettings.tsx`
- Create: `web/src/components/RiskProfileEditor.tsx`
- Create: `web/src/components/ModelRoutingEditor.tsx`
- Create: `web/src/hooks/useRiskProfile.ts`

- [ ] **Step 1: Create useRiskProfile hook**

TanStack Query + mutation hooks for `GET/PUT /api/projects/:id/risk-profile`.

- [ ] **Step 2: Create RiskProfileEditor component**

Edit what qualifies as low/medium/high risk. JSON editor or structured form.

- [ ] **Step 3: Create ModelRoutingEditor component**

Configure which model tier each risk level maps to. Dropdown selectors: low → [haiku, sonnet, opus], medium → [...], high → [...]. Show estimated cost impact when changing.

- [ ] **Step 4: Create ProjectSettings page**

Combines RiskProfileEditor, ModelRoutingEditor, and resource threshold settings. Accessible from the project dashboard.

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/ProjectSettings.tsx web/src/components/RiskProfile*.tsx web/src/components/ModelRouting*.tsx web/src/hooks/useRiskProfile.ts
git commit -m "feat: add settings page with risk profile and model routing editors"
```

**Status:** `pending`
**Branch:** `phase/ui`

---

### Task 5.8: PO chat window

**Files:**
- Create: `web/src/components/POChatWindow.tsx`
- Create: `web/src/hooks/usePOChat.ts`

- [ ] **Step 1: Create usePOChat hook**

Manages a PO chat session via the API. Sends user messages to `POST /api/projects/:id/po/chat`, receives streamed responses via WebSocket. Handles session lifecycle (open/close).

- [ ] **Step 2: Create POChatWindow component**

Slide-in panel on the project dashboard. Chat interface with message history (current session only — history is not persisted in the UI). "Close" button ends the PO session.

For planning phase: the chat window is the primary interface. Full-screen chat layout.
For execution phase: slide-in panel alongside the kanban board.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/POChatWindow.tsx web/src/hooks/usePOChat.ts
git commit -m "feat: add PO chat window for planning and execution conversations"
```

**Status:** `pending`
**Branch:** `phase/ui`

---

## Phase 6: Integration and end-to-end flow

> Depends on: Phase 5 complete
> Branch: `phase/integration`

### Task 6.1: PO workspace and session manager

> See [[PO Prompt Architecture]] for the full design.

**Files (source — deployed to `~/foundry/` at install time):**
- Create: `foundry-po/CLAUDE.md` → deploys to `~/foundry/CLAUDE.md`
- Create: `foundry-po/playbooks/planning.md` → deploys to `~/foundry/playbooks/planning.md`
- Create: `foundry-po/playbooks/estimation.md`
- Create: `foundry-po/playbooks/review.md`
- Create: `foundry-po/playbooks/execution-chat.md`
- Create: `foundry-po/playbooks/escalation.md`
- Create: `foundry-po/playbooks/phase-transition.md`
- Create: `internal/po/session.go`
- Create: `internal/po/context.go`
- Create: `internal/po/session_test.go`

`foundry-po/` is the source directory in the repo. At setup/install time, these files are copied to `~/foundry/` on the host machine where the PO runs. The `internal/po/` Go package manages sessions and handles the deployment.

- [ ] **Step 1: Create PO base CLAUDE.md**

The PO's identity and workspace conventions. Loaded automatically every session via Claude Code's native CLAUDE.md loading. Contains: role definition, workspace layout, session startup sequence, file writing conventions. See [[PO Prompt Architecture]] for exact content.

- [ ] **Step 2: Create all six playbooks**

One markdown file per session type. Each playbook is self-contained: what to read on startup, how to behave, what to write as output. See [[PO Prompt Architecture]] for exact content of each playbook.

Key: the planning playbook must include the `/grill-me` skill integration — heavily encouraged before spec approval, with behavioral consequences if skipped.

- [ ] **Step 3: Define PO workspace file schemas**

The PO reads and writes these files. Define JSON schemas for the structured ones so the control plane can parse PO output:

`project.yaml`:
```yaml
name: soapbox
repo: github.com/Quantum-Drift/soapbox
tech_stack: [go, react, postgres]
created: 2026-03-23
```

`team.json`:
```json
{
  "roles": ["backend-developer", "frontend-developer"],
  "agent_count": 2,
  "token_estimate": 150000,
  "cost_estimate_usd": 12.50
}
```

`estimate.json`:
```json
{
  "phases": [
    {"name": "Phase 1", "tasks": 5, "token_estimate": 50000, "cost_estimate_usd": 4.00}
  ],
  "total_tokens": 150000,
  "total_cost_usd": 12.50
}
```

Include schemas in the estimation playbook so the PO knows the exact format.

- [ ] **Step 4: Implement POSessionManager**

Go service that launches and manages PO Claude Code sessions as local subprocesses. The PO runs on the host machine, NOT in a container.

```go
type POSessionManager struct {
    foundryHome string // ~/foundry/
    apiKey      string
}

// StartSession launches a PO session with the given context
func (m *POSessionManager) StartSession(ctx context.Context, opts POSessionOpts) (*POSession, error)

// BuildSessionContext assembles the --append-system-prompt block
func (m *POSessionManager) BuildSessionContext(opts POSessionOpts) string
```

- [ ] **Step 5: Implement session context builder**

Builds the structured `[foundry:session]` block from session options. Reads `project.yaml` for tech stack and repo info. Maps UI actions to session types and playbooks.

```go
type POSessionOpts struct {
    Type      string // planning, estimation, review, execution-chat, escalation, phase-transition
    Project   string // project name
    Trigger   string // user, system
    // System-triggered fields:
    TaskID    string
    TaskTitle string
    AgentRole string
    RiskLevel string
    Branch    string
}
```

- [ ] **Step 6: Implement PO session launcher**

Uses Claude Code CLI to launch sessions. Model tier varies by session type (see [[PO Prompt Architecture]] for the full mapping):
- Planning/estimation (user-triggered, Opus): `claude -p <message> --model opus --output-format stream-json --max-turns 50 --append-system-prompt <context>`
- Review/escalation/phase-transition (system-triggered, Sonnet): `claude --bare -p <prompt> --model sonnet --output-format stream-json --max-budget-usd 2.00 --max-turns 20 --dangerously-skip-permissions --append-system-prompt <context>`
- Execution chat (user-triggered, Sonnet): `claude -p <message> --model sonnet --output-format stream-json --max-turns 30 --append-system-prompt <context>`

The session manager resolves the model from a session-type-to-model mapping, not a hardcoded value.

Parse NDJSON output to extract events, token usage, and session results. Forward events to the WebSocket hub for UI display.

- [ ] **Step 7: Scaffold a default project workspace**

When a new project is created, the control plane creates the project directory with the initial structure:

```
~/foundry/projects/<name>/
├── project.yaml    ← populated from project creation form
├── memory/         ← empty
├── decisions/      ← empty
└── artifacts/      ← empty
```

- [ ] **Step 8: Run tests**

```bash
go test ./internal/po/ -v
```

- [ ] **Step 9: Commit**

```bash
git add foundry-po/ internal/po/
git commit -m "feat: add PO workspace, playbooks, and session manager"
```

**Status:** `pending`
**Branch:** `phase/integration`

---

### Task 6.2: Project start flow

**Files:**
- Create: `internal/orchestrator/starter.go`
- Create: `internal/orchestrator/starter_test.go`

- [ ] **Step 1: Write failing test for project start sequence**

Test the full flow: read team composition from spec, create agent records in DB, spin up containers, assign initial tasks.

- [ ] **Step 2: Implement ProjectStarter**

When a project is approved and the user hits "Start":
1. Read `team.json` and `estimate.json` from the PO workspace (`~/foundry/projects/<name>/`)
2. Read `plan.md` to extract tasks with dependencies and risk levels
3. Create Agent records in Postgres for each role in the team
4. Resolve model tiers for each task using TierResolver (risk_level → model_routing → tier)
5. Create the Docker network for the project
6. Create the team container via ContainerManager.CreateTeam() — one container, all agents inside
7. Create Task records in Postgres (with resolved model_tier per task)
8. Run DAG resolver to find initial unblocked tasks
9. Publish initial task assignments via RabbitMQ to the team container's sidecar
10. Update project status to `active`

- [ ] **Step 3: Run tests**

```bash
go test ./internal/orchestrator/ -v -run TestStarter
```

- [ ] **Step 4: Commit**

```bash
git add internal/orchestrator/starter.go internal/orchestrator/starter_test.go
git commit -m "feat: add project start flow with team spin-up and task assignment"
```

**Status:** `pending`
**Branch:** `phase/integration`

---

### Task 6.3: Clean pause flow

**Files:**
- Create: `internal/orchestrator/pause.go`
- Create: `internal/orchestrator/pause_test.go`

- [ ] **Step 1: Write failing test for clean agent pause**

Test that pausing an agent: sends command, waits for context summary, stops container, updates DB.

- [ ] **Step 2: Implement clean pause**

Single agent pause:
1. Send `pause` command to agent via RabbitMQ
2. Agent receives command, finishes current subtask
3. Agent commits work, writes context summary to `/shared/context/<agent_id>.md`
4. Agent publishes `agent_paused` event
5. Go receives event, stops the container, updates agent status to `paused`
6. Updates task status to `paused`

Team pause: iterates all active agents, pauses each one. Updates project status to `paused`.

Resume: starts container, reassigns paused task, updates statuses.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/orchestrator/ -v -run TestPause
```

- [ ] **Step 4: Commit**

```bash
git add internal/orchestrator/pause.go internal/orchestrator/pause_test.go
git commit -m "feat: add clean pause/resume flow for agents and teams"
```

**Status:** `pending`
**Branch:** `phase/integration`

---

### Task 6.4: End-to-end smoke test

**Files:**
- Create: `tests/e2e/smoke_test.go`

- [ ] **Step 1: Write an end-to-end test**

Test the full lifecycle with a simple project:
1. Create a project via API
2. Verify PO local process starts
3. Simulate PO writing spec and tasks to shared volume
4. Approve the spec via API
5. Start the project via API
6. Verify the team container is running and agents are active
7. Simulate a task completion event
8. Verify DAG resolver assigns the next task
9. Pause an agent, verify clean shutdown
10. Resume, verify task is reassigned

This test uses the real Docker daemon, Postgres, Redis, and RabbitMQ (from docker-compose).

- [ ] **Step 2: Run the smoke test**

```bash
make docker-up
go test ./tests/e2e/ -v -timeout 120s
```

- [ ] **Step 3: Commit**

```bash
git add tests/e2e/
git commit -m "test: add end-to-end smoke test for full project lifecycle"
```

**Status:** `pending`
**Branch:** `phase/integration`

---

### Task 6.5: Frontend integration verification

- [ ] **Step 1: Start all services**

```bash
make docker-up
make run &
cd web && npm run dev
```

- [ ] **Step 2: Manual verification checklist**

- Create a project in the UI
- View the project in the project list
- Open the project dashboard
- Verify the kanban board renders
- Verify agent status bar renders
- Verify WebSocket connection establishes
- Navigate to agent detail page
- Verify log stream connects

- [ ] **Step 3: Fix any integration issues found**

- [ ] **Step 4: Commit any fixes**

```bash
git add -A
git commit -m "fix: resolve frontend integration issues from smoke testing"
```

**Status:** `pending`
**Branch:** `phase/integration`

---

## Phase 7: Local mode (optional)

> Depends on: Phase 6 complete (full Docker mode working end-to-end)
> Branch: `phase/local-mode`
> Note: This phase is optional. Foundry is fully functional after Phase 6 using Docker mode. Local mode removes the Docker requirement by running agents as local processes with folder-based isolation and subscription auth.

### Task 7.1: LocalRuntime implementation

**Files:**
- Create: `internal/runtime/local.go`
- Create: `internal/runtime/local_test.go`

- [ ] **Step 1: Write failing tests for LocalRuntime.Setup**

Test that Setup creates the project folder structure: `workspace/`, `worktrees/`, `shared/` (with subdirectories), `state/`. Test that it clones the repo into `workspace/`.

- [ ] **Step 2: Implement LocalRuntime.Setup**

Implements the `Runtime` interface without Docker. Creates the project directory at `<project_dir>/<name>/`, clones the repo, creates `shared/{contracts,designs,context,reviews,messages,status}` and `state/` directories.

- [ ] **Step 3: Write failing tests for LocalRuntime.LaunchAgent**

Test that LaunchAgent creates a git worktree, generates composite CLAUDE.md, spawns `claude -p` with the right flags, and registers the process in `state/agents.json`. Test concurrency limiting — launching beyond `max_concurrent_agents` should block.

- [ ] **Step 4: Implement LocalRuntime.LaunchAgent**

```go
func (r *LocalRuntime) LaunchAgent(ctx context.Context, opts AgentOpts) (AgentProcess, error) {
    r.sem <- struct{}{} // acquire concurrency slot

    // Create worktree if needed
    // Generate composite CLAUDE.md
    // Spawn agent process
    cmd := exec.CommandContext(ctx, "claude",
        "-p", opts.Prompt,
        "--output-format", "stream-json",
        "--model", opts.Model,
        "--max-turns", strconv.Itoa(opts.MaxTurns),
        "--dangerously-skip-permissions",
        "--allowedTools", strings.Join(opts.AllowedTools, ","),
    )
    cmd.Dir = opts.WorkDir
    // ...
}
```

No `--bare` — local mode uses subscription auth. No `--max-budget-usd` — subscription, not API.

- [ ] **Step 5: Write failing tests for LocalRuntime.WatchEvents**

Test that writing a file to `shared/status/` triggers an event on the channel. Test that writing to `shared/messages/` triggers a message event.

- [ ] **Step 6: Implement LocalRuntime.WatchEvents**

Uses fsnotify to watch `shared/` recursively. Parses JSON files on change, classifies by subdirectory (status → completion event, messages → message event, contracts → artifact event), sends typed events on the channel. No RabbitMQ needed.

- [ ] **Step 7: Implement LocalRuntime.StopAgent and Cleanup**

StopAgent writes `pause-signal` to `state/`, waits for the process to exit (with timeout), then releases the concurrency slot. Cleanup removes the project folder on project completion.

- [ ] **Step 8: Run tests**

```bash
go test ./internal/runtime/ -v -run TestLocal
```

- [ ] **Step 9: Commit**

```bash
git add internal/runtime/local.go internal/runtime/local_test.go
git commit -m "feat: add LocalRuntime with fsnotify event watching and folder-based isolation"
```

**Status:** `pending`
**Branch:** `phase/local-mode`

---

### Task 7.2: Runtime mode configuration

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Modify: `cmd/foundry/main.go`

- [ ] **Step 1: Add runtime_mode to config**

Add `FOUNDRY_RUNTIME_MODE` env var (values: `docker` (default), `local`). When `local`, RabbitMQ connection is optional.

- [ ] **Step 2: Update composition root**

Initialize the appropriate Runtime (LocalRuntime or DockerRuntime based on config). Skip RabbitMQ connection when in local mode.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/config/ -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/config/ cmd/foundry/main.go
git commit -m "feat: add runtime mode config with local/docker switching"
```

**Status:** `pending`
**Branch:** `phase/local-mode`

---

### Task 7.3: Local mode event router and log streaming

**Files:**
- Modify: `internal/event/router.go`
- Create: `internal/event/local_router.go`
- Create: `internal/event/local_router_test.go`

- [ ] **Step 1: Write failing test for local log streaming**

Test that agent stdout from local processes is captured and forwarded to the WebSocket hub.

- [ ] **Step 2: Implement local event routing**

For local mode, the event router reads stdout from agent processes directly (via `cmd.StdoutPipe()`) instead of subscribing to RabbitMQ. Forwards log lines to WebSocket hub (scoped to agent ID) and batch-persists to Postgres for history.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/event/ -v -run TestLocal
```

- [ ] **Step 4: Commit**

```bash
git add internal/event/
git commit -m "feat: add local mode event router with stdout-based log streaming"
```

**Status:** `pending`
**Branch:** `phase/local-mode`

---

### Task 7.4: Local mode PO sessions

**Files:**
- Modify: `internal/po/session.go`

- [ ] **Step 1: Update PO session launcher for local mode**

In local mode, PO sessions use subscription auth instead of API keys:
- Remove `--bare` flag (uses subscription auth)
- Remove `--max-budget-usd` (subscription, not API)
- Keep all other flags the same

- [ ] **Step 2: Run tests**

```bash
go test ./internal/po/ -v
```

- [ ] **Step 3: Commit**

```bash
git add internal/po/
git commit -m "feat: add local mode PO session support with subscription auth"
```

**Status:** `pending`
**Branch:** `phase/local-mode`

---

### Task 7.5: Local mode e2e test

**Files:**
- Create: `tests/e2e/local_smoke_test.go`

- [ ] **Step 1: Write an end-to-end test for local mode**

Test the full lifecycle with `runtime_mode: local`:
1. Create a project via API
2. Verify PO local process starts
3. Simulate PO writing spec and tasks
4. Approve the spec via API
5. Start the project via API
6. Verify project folder is created and agents are spawned as local processes
7. Simulate a task completion event (write status file to `shared/status/`)
8. Verify fsnotify picks it up and DAG resolver assigns the next task
9. Pause an agent, verify clean shutdown
10. Resume, verify task is reassigned

This test uses Postgres and Redis (from docker-compose). No Docker daemon or RabbitMQ needed.

- [ ] **Step 2: Run the smoke test**

```bash
make docker-up
go test ./tests/e2e/ -v -run TestLocalSmoke -timeout 120s
```

- [ ] **Step 3: Commit**

```bash
git add tests/e2e/local_smoke_test.go
git commit -m "test: add local mode end-to-end smoke test"
```

**Status:** `pending`
**Branch:** `phase/local-mode`

---

## Dependency graph

```mermaid
graph LR
    P0[Phase 0<br/>Bootstrap<br/>10 tasks] --> P1[Phase 1<br/>Data Layer<br/>4 tasks]
    P1 --> P2[Phase 2<br/>Containers<br/>5 tasks]
    P2 --> P3[Phase 3<br/>Orchestration +<br/>Token Optimization<br/>6 tasks]
    P3 --> P4[Phase 4<br/>Project API +<br/>PO API<br/>8 tasks]
    P4 --> P5[Phase 5<br/>React UI +<br/>Token Dashboard +<br/>PO Chat<br/>8 tasks]
    P5 --> P6[Phase 6<br/>Integration +<br/>PO Workspace<br/>5 tasks]
    P6 -.->|optional| P7[Phase 7<br/>Local Mode<br/>5 tasks]
```

Phases 0-6 are sequential and deliver a fully working Foundry with Docker mode. Phase 7 is optional and adds local runtime support (folder-based isolation, fsnotify, subscription auth). Within each phase, tasks are generally sequential (each builds on the previous), except where noted.

## Summary

- **Docker runtime mode** (Phases 0-6): Docker is the primary runtime. Agents run in team containers with RabbitMQ for communication, sidecar for event routing, and API key auth. Phases 0-6 deliver a fully working Foundry.
- **Local runtime mode** (Phase 7, optional): LocalRuntime implements the same Runtime interface. Agents run as local processes using subscription auth, no Docker or RabbitMQ needed. Folder-based isolation with fsnotify for event detection.
- **Token optimization** (Phase 3): Model tier resolution, response caching, budget tracking, price table. Risk level drives model selection — low→haiku, medium→sonnet, high→opus.
- **PO prompt architecture** (Phase 6): PO runs as a Claude Code process. Stateless sessions with playbook-based behavior. Six playbooks for six session types. Context injected via `--append-system-prompt`.
- **One container per team** (Phase 2): Single Docker container per project with process supervisor, RabbitMQ sidecar, and git worktrees for branch isolation.
- **Data model** (Phase 1): Task has `model_tier`, `token_usage`, `automation_eligible`. RiskProfile has `model_routing`. Tables: `risk_profiles`, `spec_mutations`, `agent_messages`.
- **UI** (Phase 5): Token dashboard, model routing settings, PO chat window.
- **PO chat API** (Phase 4): Endpoints for PO chat, planning, estimation, and session management.
