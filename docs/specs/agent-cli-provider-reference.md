---
tags:
  - type/resource
  - domain/tech
  - project/foundry
created: 2026-03-23
status: growing
---

# Agent CLI provider reference

Implementation reference for Foundry's agent provider abstraction. Covers the three MVP-track CLIs: Claude Code, Gemini CLI, and Codex CLI. Organized by cross-cutting concern so an implementer can see how each provider handles the same problem.

Related: [[Design Specification]], [[Implementation Plan]]

Detailed per-provider docs: [[Claude Code CLI Programmatic Interface]], [[Gemini CLI Integration Reference]], [[Research — OpenAI Codex CLI Programmatic Interface]]

## Process model

All three CLIs follow the same fundamental pattern: spawn a subprocess, pass a prompt, read structured output from stdout, wait for exit. Each invocation runs a full agentic loop internally (the agent makes multiple tool calls, reads files, runs commands) before exiting. None support injecting messages into a running session.

| Concern | Claude Code | Gemini CLI | Codex CLI |
|---------|-------------|------------|-----------|
| Headless command | `claude --bare -p "task"` | `gemini -p "task"` | `codex exec "task"` |
| Prompt delivery | `-p` flag or stdin pipe | `-p` flag or stdin pipe | Positional arg or stdin pipe |
| Mid-session messages | Not possible (use `--resume`) | Not possible (use `-r`) | Not possible (use `exec resume`) |
| Session resume | `--resume <session_id>` | `-r <session_id>` | `exec resume <thread_id>` |
| Structured output | `--output-format stream-json` | `--output-format stream-json` | `--json` |
| Graceful stop | SIGTERM, wait 5s, SIGKILL | SIGTERM/SIGINT | SIGTERM |
| Process isolation | Fully independent, no locks | Fully independent, no locks | Fully independent, no locks |

### Foundry implication

Each agent task maps to one subprocess invocation. The Go control plane spawns the process, reads NDJSON from stdout, and waits for exit. For multi-step tasks where the PO needs to give follow-up instructions, the control plane stops the current session and starts a new one with `--resume` / `-r` / `exec resume`, passing the session/thread ID from the previous run.

All three CLIs support running multiple processes simultaneously with no shared state or lock files. Concurrency is safe at the process level.

## Invocation patterns

### Claude Code

```bash
claude --bare -p "task prompt" \
  --output-format stream-json \
  --model sonnet \
  --max-budget-usd 5.00 \
  --max-turns 30 \
  --dangerously-skip-permissions \
  --allowedTools "Read,Edit,Bash,Glob,Grep" \
  --append-system-prompt "Follow Go conventions. You are a backend developer." \
  --no-session-persistence
```

Key flags:
- `--bare` skips all auto-discovery (CLAUDE.md, hooks, MCP, keychain). Fast, deterministic startup. Auth comes strictly from `ANTHROPIC_API_KEY`. This is the right mode for containers.
- `--dangerously-skip-permissions` auto-approves all tool use. Only safe inside sandboxed containers.
- `--allowedTools` restricts which tools the agent can use (still available but require permission if not listed).
- `--append-system-prompt` injects role-specific instructions without replacing the default system prompt.
- `--no-session-persistence` prevents writing session files inside containers.
- `--max-budget-usd` and `--max-turns` provide built-in cost and runaway protection.

### Gemini CLI

```bash
gemini -p "task prompt" \
  --output-format stream-json \
  --approval-mode yolo \
  -m gemini-2.5-pro
```

Key flags:
- `-p` forces non-interactive mode.
- `--approval-mode yolo` auto-approves all tool calls. Without this, the process hangs waiting for user input.
- `-m` selects the model.
- No equivalent to `--bare` — Gemini always loads GEMINI.md from the working directory. Write your agent instructions there.
- No built-in budget cap. Turn limit via `model.maxSessionTurns` in settings.json only.

### Codex CLI

```bash
codex exec --json --full-auto --ephemeral \
  --model o4-mini \
  --cd /worktree/path \
  --skip-git-repo-check \
  "task prompt"
```

Key flags:
- `exec` is the non-interactive subcommand.
- `--json` enables JSONL event streaming on stdout.
- `--full-auto` sets `approval_policy=on-request` + `sandbox=workspace-write`. Agent can read/write files and run commands within the working directory.
- `--ephemeral` prevents session persistence to disk. Use this for throwaway tasks.
- `--cd` sets the working directory.
- `--skip-git-repo-check` needed if the worktree isn't a full git repo.
- Prompt is passed via stdin: `cmd.Stdin = strings.NewReader(prompt)`.
- No built-in budget cap. Foundry must track `turn.completed` usage events.

Alternative: `--yolo` (alias `--dangerously-bypass-approvals-and-sandbox`) when Foundry provides its own Docker sandbox. Conflicts with `--full-auto`.

## Go subprocess pattern

All three providers follow the same Go pattern:

```go
func (p *Provider) Start(ctx context.Context, opts SessionOpts) (Session, error) {
    args := p.buildArgs(opts) // provider-specific flag construction
    cmd := exec.CommandContext(ctx, p.binary, args...)
    cmd.Dir = opts.WorkDir
    cmd.Env = p.buildEnv(opts) // provider-specific env vars

    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()

    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("start %s: %w", p.Name(), err)
    }

    return &session{
        cmd:    cmd,
        stdout: stdout,
        stderr: stderr,
        events: make(chan Event, 64),
    }, nil
}
```

The `session.Output()` method returns a channel that a goroutine populates by scanning NDJSON from stdout:

```go
func (s *session) readLoop() {
    scanner := bufio.NewScanner(s.stdout)
    for scanner.Scan() {
        var raw json.RawMessage
        if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
            continue
        }
        event := s.provider.ParseEvent(raw) // provider-specific parsing
        s.events <- event
    }
    close(s.events)
}
```

Each provider implements `ParseEvent()` to normalize its CLI's event format into Foundry's internal `Event` type.

## Output event formats

All three emit NDJSON (one JSON object per line). The event schemas differ but map to the same logical phases: init, tool use, tool result, progress, and final result with usage stats.

### Claude Code events

```
type: "system"    subtype: "init"              → session started
type: "system"    subtype: "api_retry"         → rate limit / retry
type: "system"    subtype: "compact_boundary"  → context compaction
type: "assistant"                              → agent response (may contain tool_use blocks)
type: "user"                                   → tool results
type: "result"                                 → final result with cost/usage
```

### Gemini CLI events

```
type: "init"         → session started, includes session_id and model
type: "message"      → user or assistant text (role field distinguishes)
type: "tool_use"     → agent requesting a tool call
type: "tool_result"  → tool execution outcome
type: "error"        → non-fatal warnings and errors
type: "result"       → final result with stats
```

### Codex CLI events

```
type: "thread.started"   → first event, contains thread_id
type: "turn.started"     → agent begins processing
type: "item.started"     → new item (command, file change, etc.)
type: "item.updated"     → item state changed
type: "item.completed"   → item reached terminal state
type: "turn.completed"   → agent finished, includes token usage
type: "turn.failed"      → turn ended with error
type: "error"            → unrecoverable stream error
```

### Normalized event type

Foundry's internal event type should normalize across providers:

```go
type EventKind string
const (
    EventInit       EventKind = "init"
    EventToolCall   EventKind = "tool_call"
    EventToolResult EventKind = "tool_result"
    EventProgress   EventKind = "progress"   // assistant text, partial output
    EventRetry      EventKind = "retry"      // rate limit, API retry
    EventResult     EventKind = "result"     // final result with usage
    EventError      EventKind = "error"
)

type Event struct {
    Kind      EventKind
    Raw       json.RawMessage // original provider event
    SessionID string          // provider's session/thread ID
    Usage     *TokenUsage     // populated on result events
    Error     *EventError     // populated on error events
    Timestamp time.Time
}

type TokenUsage struct {
    InputTokens  int
    OutputTokens int
    CachedTokens int
    TotalCostUSD float64 // only Claude reports this directly
}
```

## Token usage extraction

This is where the providers diverge most. Each reports usage differently, and Foundry needs a consistent view for budget tracking.

### Claude Code

Usage is reported in the final `result` event:

```json
{
  "type": "result",
  "total_cost_usd": 0.0994,
  "usage": {
    "input_tokens": 3,
    "cache_creation_input_tokens": 15325,
    "cache_read_input_tokens": 6692,
    "output_tokens": 11
  },
  "modelUsage": {
    "claude-sonnet-4-6": {
      "inputTokens": 3,
      "outputTokens": 11,
      "cacheReadInputTokens": 6692,
      "cacheCreationInputTokens": 15325,
      "costUSD": 0.0994
    }
  }
}
```

Claude is the only provider that reports `total_cost_usd` directly. The `modelUsage` breakdown shows per-model consumption when multiple models were used (e.g., main model + subagent model).

### Gemini CLI

Usage is in the `result` event's `stats` field:

```json
{
  "type": "result",
  "stats": {
    "total_tokens": 15000,
    "input_tokens": 12000,
    "output_tokens": 3000,
    "cached": 8000,
    "input": 4000,
    "duration_ms": 45000,
    "tool_calls": 5,
    "models": {
      "gemini-2.5-pro": {
        "total_tokens": 15000,
        "input_tokens": 12000,
        "output_tokens": 3000,
        "cached": 8000
      }
    }
  }
}
```

No cost reported. Foundry must calculate cost from token counts and a price table. The `models` breakdown shows per-model usage.

### Codex CLI

Usage is reported per-turn on `turn.completed` events:

```json
{
  "type": "turn.completed",
  "usage": {
    "input_tokens": 24763,
    "cached_input_tokens": 24448,
    "output_tokens": 122
  }
}
```

No cost reported. No per-model breakdown. If a session has multiple turns (via resume), Foundry accumulates usage across `turn.completed` events.

### Cost calculation

Only Claude reports cost directly. For Gemini and Codex, Foundry maintains a price table:

```go
type ModelPricing struct {
    InputPer1K  float64
    OutputPer1K float64
    CachedPer1K float64
}

var PriceTable = map[string]ModelPricing{
    // Claude
    "haiku":  {InputPer1K: 0.00025, OutputPer1K: 0.00125, CachedPer1K: 0.0000625},
    "sonnet": {InputPer1K: 0.003,   OutputPer1K: 0.015,   CachedPer1K: 0.00075},
    "opus":   {InputPer1K: 0.015,   OutputPer1K: 0.075,   CachedPer1K: 0.00375},
    // Gemini
    "gemini-2.5-flash":     {InputPer1K: 0.00015, OutputPer1K: 0.0006, CachedPer1K: 0.0000375},
    "gemini-2.5-pro":       {InputPer1K: 0.00125, OutputPer1K: 0.01,   CachedPer1K: 0.000315},
    // OpenAI
    "o4-mini":    {InputPer1K: 0.0011, OutputPer1K: 0.0044, CachedPer1K: 0.000275},
    "gpt-5.4":    {InputPer1K: 0.005,  OutputPer1K: 0.015,  CachedPer1K: 0.00125},
}
```

Price table lives in config, not code. Updated when providers change pricing.

## Model selection

### Tier-to-model mapping

| Abstract tier | Claude Code | Gemini CLI | Codex CLI |
|---------------|------------|------------|-----------|
| `haiku` | `haiku` | `gemini-2.5-flash` | `o4-mini` |
| `sonnet` | `sonnet` | `gemini-2.5-pro` | `gpt-5.4` |
| `opus` | `opus` | `gemini-2.5-pro` | `gpt-5.4` |

Claude accepts tier aliases directly (`--model haiku`). Gemini and Codex require concrete model names.

Note: Gemini and Codex have fewer tiers than Claude. `gemini-2.5-pro` covers both sonnet and opus equivalent. `gpt-5.4` is the top tier for Codex. The mapping config allows users to adjust this as new models ship.

### Model flag per CLI

| CLI | Flag | Example |
|-----|------|---------|
| Claude Code | `--model <alias or full ID>` | `--model haiku`, `--model claude-sonnet-4-6` |
| Gemini CLI | `-m <model name>` | `-m gemini-2.5-flash`, `-m gemini-2.5-pro` |
| Codex CLI | `--model <name>` or `-m <name>` | `-m o4-mini`, `-m gpt-5.4` |

### Effort/reasoning controls

Some providers support adjusting reasoning depth independently of model selection:

| CLI | Flag | Values |
|-----|------|--------|
| Claude Code | `--effort` | `low`, `high`, `max` (Opus only) |
| Gemini CLI | None | N/A |
| Codex CLI | `--config model_reasoning_effort=` | `minimal`, `low`, `medium`, `high`, `xhigh` |

Foundry can use effort levels as a secondary cost lever. Low-risk tasks on Haiku-class models can also use low effort. High-risk tasks on Opus-class models can use high effort.

## Authentication

### Environment variables

| CLI | Primary env var | Notes |
|-----|----------------|-------|
| Claude Code | `ANTHROPIC_API_KEY` | Used instead of any subscription when set. Multiple processes share the same key. |
| Gemini CLI | `GEMINI_API_KEY` | Free tier: 10 RPM. Paid tier: higher limits. Alternatively, pre-auth via Google OAuth for 60 RPM free. |
| Codex CLI | `CODEX_API_KEY` | Also reads `OPENAI_API_KEY`. Standard OpenAI rate limits apply. |

### Rate limits (concurrent agent implications)

| Provider | Free tier | Paid tier | Shared across |
|----------|----------|-----------|---------------|
| Claude (API key) | Per org tier limits | Per org tier limits | All processes using same key |
| Gemini (API key) | 10 RPM, 250 RPD | Higher, per-project | All processes using same project |
| Gemini (OAuth) | 60 RPM, 1000 RPD | Per-project | All processes using same credentials |
| Codex | Standard OpenAI limits | Standard OpenAI limits | All processes using same key |

For a Foundry project with 5 concurrent agents, Claude's API key model works well (rate limits are generous on paid tiers). Gemini's free API key tier (10 RPM) is tight for multi-agent use — OAuth or paid tier recommended. Codex follows standard OpenAI rate limits.

The Go control plane should track request rates per provider and throttle agent launches if approaching limits. Rate limit errors surface as retry events (Claude: `system/api_retry`, Gemini: `error` events, Codex: `turn.failed`).

## Permissions and sandboxing

| Concern | Claude Code | Gemini CLI | Codex CLI |
|---------|-------------|------------|-----------|
| Auto-approve all | `--dangerously-skip-permissions` | `--approval-mode yolo` | `--full-auto` or `--yolo` |
| Tool restriction | `--allowedTools "Read,Edit,Bash"` | Settings-based policy engine | None (full access or nothing) |
| Sandbox mode | None (relies on container) | `--sandbox docker` or `runsc` | `--sandbox workspace-write` or container |
| Instruction injection | `--append-system-prompt` | Write to GEMINI.md in workdir | Write to AGENTS.md in workdir |
| Skip auto-discovery | `--bare` | No equivalent | `--skip-git-repo-check` |

### Foundry's approach

Since Foundry agents run inside Docker containers, the container is the security boundary. All three CLIs should run in their most permissive mode:

- Claude: `--bare --dangerously-skip-permissions`
- Gemini: `--approval-mode yolo`
- Codex: `--yolo` (since Foundry's container provides sandboxing)

Tool restriction is handled at the Foundry level through the agent role definition, not at the CLI level. Claude's `--allowedTools` is a nice defense-in-depth layer that Foundry should use when available.

## Context and instruction injection

Each CLI loads instructions from files in the working directory. Foundry controls what instructions each agent sees by writing the appropriate files to each worktree before launching the agent.

| CLI | Instruction file | Loading behavior |
|-----|-----------------|------------------|
| Claude Code | `CLAUDE.md` + `.claude/` directory | Hierarchical: `~/.claude/`, project root, current dir. `--bare` skips all. |
| Gemini CLI | `GEMINI.md` + `.gemini/` directory | Hierarchical: `~/.gemini/`, project root, current dir. Supports `@file.md` imports. |
| Codex CLI | `AGENTS.md` | Merges: `~/.codex/AGENTS.md`, repo root, current dir. `--no-project-doc` skips. |

### Foundry's approach

During worktree setup, the entrypoint script writes provider-appropriate instruction files:

1. **Claude**: Composite `CLAUDE.md` (base + project overlay + role definition) placed in the worktree root. `.claude/` directory with language/framework files. `--bare` is used so only the worktree's files are loaded.
2. **Gemini**: Equivalent `GEMINI.md` placed in the worktree root. `.gemini/settings.json` for MCP and tool config.
3. **Codex**: `AGENTS.md` placed in the worktree root. `~/.codex/config.toml` for model and approval settings.

The content of these files is identical in substance (role definition, coding standards, communication protocol), just formatted for each CLI's conventions.

## Session management

| Concern | Claude Code | Gemini CLI | Codex CLI |
|---------|-------------|------------|-----------|
| Session ID field | `session_id` (in init and result) | `session_id` (in init event) | `thread_id` (in `thread.started`) |
| Resume flag | `--resume <id>` | `-r <id>` | `exec resume <id> "prompt"` |
| Disable persistence | `--no-session-persistence` | N/A (sessions in `~/.gemini/tmp/`) | `--ephemeral` |
| Storage location | `~/.claude/projects/<hash>/<id>.jsonl` | `~/.gemini/tmp/<hash>/chats/` | `~/.codex/sessions/` (SQLite) |

### Foundry's approach

For single-turn tasks (most common), disable session persistence to avoid disk clutter in containers:
- Claude: `--no-session-persistence`
- Codex: `--ephemeral`
- Gemini: no flag needed, sessions are lightweight

For multi-turn tasks (PO follow-ups), capture the session/thread ID from the first event and pass it to subsequent invocations. The control plane stores these IDs in Postgres alongside the task record.

## Cost control

| Mechanism | Claude Code | Gemini CLI | Codex CLI |
|-----------|-------------|------------|-----------|
| Budget cap | `--max-budget-usd 5.00` | None | None |
| Turn limit | `--max-turns 30` | `model.maxSessionTurns` in settings | None |
| Effort control | `--effort low/high/max` | None | `--config model_reasoning_effort=` |
| Token limit | `CLAUDE_CODE_MAX_OUTPUT_TOKENS` env | `model.summarizeToolOutput` in settings | None |

Only Claude has built-in budget caps. For Gemini and Codex, Foundry must:

1. Track token usage from result/completion events
2. Calculate cost using the price table
3. Kill the process with SIGTERM if the task budget is exceeded

The control plane should always set Claude's `--max-budget-usd` as a per-task safety net, even when Foundry is also tracking budget independently. Belt and suspenders.

## Exit codes

| Code | Claude Code | Gemini CLI | Codex CLI |
|------|-------------|------------|-----------|
| 0 | Success | Success | Success |
| 1 | Error (check `subtype` for detail) | General error | General error |
| 42 | N/A | Input error | N/A |
| 53 | N/A | Turn limit exceeded | N/A |
| 130 | N/A | User interrupt (SIGINT) | N/A |
| 137 | N/A | N/A | Killed by SIGKILL |

Claude Code's `result` event includes a `subtype` field for granular status:
- `success` — completed normally
- `error_max_turns` — hit turn limit
- `error_max_budget_usd` — hit budget limit
- `error_during_execution` — API failure or cancellation

## Known issues and workarounds

### Claude Code
- **No known critical bugs for headless mode.** The `--bare` flag provides clean, predictable behavior.
- Auto-retry on rate limits with `system/api_retry` events. The control plane should log these but doesn't need to intervene.

### Gemini CLI
- **Headless hang on tool permission** (issue #19774, open): The agent enters an infinite loop when tools aren't available instead of failing. Always use `--approval-mode yolo` and wrap all invocations in `context.WithTimeout`.
- **Rate limit infinite retry** (issues #1626, #1631): Can get stuck retrying rate-limited requests. The control plane needs a watchdog timeout.
- **OAuth in headless**: Requires interactive browser auth on first use. Use `GEMINI_API_KEY` for containers.
- **Sub-agent model override**: The `-m` flag only controls the main agent. Sub-agents may use different models.

### Codex CLI
- **No built-in budget caps.** Foundry must implement its own cost tracking.
- **Linux sandbox limitations.** No sandboxing by default on Linux. Use `--yolo` when running inside Foundry's Docker containers.
- **Git repo requirement.** Default behavior requires a git repo. Use `--skip-git-repo-check` in worktrees that aren't full repos.

### Cross-provider watchdog

Given the known hang bugs in Gemini and the potential for any CLI to get stuck, every agent session should run with a deadline:

```go
timeout := time.Duration(task.EstimatedMinutes*2) * time.Minute
if timeout < 5*time.Minute {
    timeout = 5 * time.Minute
}
ctx, cancel := context.WithTimeout(parentCtx, timeout)
defer cancel()

cmd := exec.CommandContext(ctx, binary, args...)
```

When the context deadline exceeds, Go kills the process with SIGKILL. The control plane logs a timeout event and notifies the PO, who decides whether to retry with a longer timeout or reassign the task.

## MCP server support

| Concern | Claude Code | Gemini CLI | Codex CLI |
|---------|-------------|------------|-----------|
| Config mechanism | `--mcp-config ./mcp.json` | `.gemini/settings.json` `mcpServers` key | Not documented for exec mode |
| Transport | stdio, SSE | stdio, SSE, HTTP streaming | N/A |
| Headless compatibility | Full support with `--bare` | stdio works, OAuth MCP needs pre-auth | N/A |

For MVP, MCP support is not required. Post-MVP, Claude and Gemini both support attaching MCP servers to headless sessions, which could enable agents to interact with external services (databases, APIs, monitoring).

## Summary: provider implementation checklist

Each provider implementation needs:

1. **`buildArgs(opts)`** — construct CLI flags from SessionOpts
2. **`buildEnv(opts)`** — set provider-specific env vars (API key, disable telemetry, etc.)
3. **`ParseEvent(raw)`** — normalize provider JSONL into Foundry's `Event` type
4. **`ModelFor(tier)`** — translate abstract tier to concrete model name
5. **`TokenUsage(resultEvent)`** — extract usage from the provider's result event format
6. **`CalculateCost(usage, model)`** — compute USD cost from token counts (Claude: use reported cost, others: price table)

The `Session` interface methods map cleanly:
- `Send()` — for multi-turn, stop current process and start new one with `--resume`
- `Output()` — channel fed by NDJSON scanner goroutine
- `Stop()` — SIGTERM the process
- `Healthy()` — check process is still running and last event was recent
