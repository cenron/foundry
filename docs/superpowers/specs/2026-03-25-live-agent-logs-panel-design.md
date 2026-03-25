# Live agent logs panel

A collapsible bottom drawer on the Project Dashboard that streams real-time agent activity during a build. Gives full visibility into what every agent is saying and doing without leaving the kanban view.

## Placement and layout

Bottom drawer on `ProjectDashboard`, below the kanban board. The dashboard layout changes from the current unconstrained flex column to a fixed-height layout (`h-[calc(100vh-<header-height>)]`) so the kanban and drawer can share the viewport. The kanban gets `flex-1 overflow-y-auto`, the drawer gets a configurable height.

A drag handle on the top edge lets the user resize the drawer height.

- **Min height:** 120px (roughly 4 log lines). Dragging below this collapses the drawer.
- **Max height:** 60% of the viewport. The kanban always keeps at least 40%.
- **Height persists** in `localStorage` across navigations and sessions.
- **Open by default** when the project status is `running`. Collapsed otherwise.
- Toggle button in the dashboard toolbar to collapse/expand.
- When collapsed and new messages arrive, the toggle shows an unread count badge. Opening the drawer resets the count to zero.

## Title bar

The drawer title bar doubles as a build progress indicator:

```
Logs — 12/18 tasks complete (67%)  [●]
```

- **Progress text:** `{completed}/{total} tasks complete ({percent}%)` — updates as tasks transition to `done`.
- **Progress bar:** subtle fill in the title bar background, left to right, tracking task completion percentage.
- **Activity indicator:** pulsing dot next to the text while any agent has a status of `working`. The `useProjectLogs` hook tracks the latest `agent.status` event per agent and derives "any active" from the set. Dot disappears when all agents are idle or the build is complete.

**Progress data source:** the task query currently owned by `KanbanBoard` gets lifted to `ProjectDashboard` and passed down to both `KanbanBoard` and `LogsTitleBar` as props. This avoids a duplicate network call.

## Log stream

Unified chronological stream of all agents' activity. No tabs, no splits.

### Visual treatment

- **Agent names** are color-coded. Each agent gets a consistent color from a fixed palette, assigned deterministically by hashing the agent ID. Colors apply to the agent name only.
- **Color palette:** 10 colors that work on both light and dark backgrounds. Defined as a constant array in `useAgentColor`. Use saturated mid-tone values (not too bright, not too dark) — similar to terminal ANSI colors. Avoid red (reserved for errors) and green (reserved for success indicators).
- **Log body text** is a single neutral color (the default text color for the theme).
- Each entry shows: `[timestamp]  AgentName  message text`

### Message tiers

Three tiers visible by default:

1. **Agent messages** — agent-to-agent and agent-to-PO communication (handoff notes, questions, status updates)
2. **Task transitions** — assignments, completions, blocked/unblocked events
3. **Errors** — agent process failures, escalations

**Agent output** (tool calls, file edits, reasoning) is collapsed by default. Each entry that has detail shows an expand chevron. Click to toggle the detail view inline.

## Interaction

- **Auto-scroll:** pinned to bottom by default. Disengages when the user scrolls up. A "jump to latest" button appears when not pinned. Buffer eviction pauses while the user is scrolled up to prevent the content shifting under them. Eviction resumes when they jump to latest or scroll back to the bottom.
- **Agent filter:** click an agent name to filter the stream to just that agent. Click again (or click a clear button) to show all agents. The filter is a quick toggle, not a modal or dropdown.
- **Expandable rows:** agent output expands/collapses inline on click.

## Data flow

### Backend

No new backend endpoints needed. Reuses the existing event pipeline — `LocalRuntime` file watcher classifies events by shared directory, `LocalRouter` processes them, and `Hub` broadcasts to all connected WebSocket clients.

Events already published cover all the message types:

| Event type | Tier | Source |
|---|---|---|
| `agent.message` | Agent messages | Files written to `shared/messages/` by agents |
| `agent.output` | Expandable detail | Default classification for unrecognized agent output |
| `task.transition` | Task transitions | Orchestrator state machine (`statemachine.go`) |
| `agent.status` | Activity indicator | Two sources: (1) files written to `shared/status/` by agents via file watcher, (2) agent stream-json output containing `{"type": "agent.status", ...}` via `LocalRouter`. If agents don't currently write status files, the activity indicator falls back to inferring activity from the presence of recent `agent.output` or `agent.message` events (any event from an agent in the last 10s = active). |

Note: there is no dedicated `agent.error` event type today. Agent errors surface as `agent.output` or `agent.status` events with error content in the payload. The `useProjectLogs` hook inspects the payload for error indicators (non-zero exit codes, error fields) and renders them in the error tier.

### Frontend

**WebSocket pattern:** follows the existing convention. `useProjectLogs` subscribes to the global `wsClient` (connected to `/ws`) and filters events client-side by `projectId` and event type. This matches how `useProjectEvents` and `useAgentLogs` already work. Migrating to per-topic ChannelHub subscriptions is a separate effort.

New hook: `useProjectLogs(projectId)` — subscribes to the global WebSocket and aggregates relevant events into a single ordered list.

- Filters events by `projectId` and the event types listed above
- Caps the buffer at 1000 entries. Older entries are dropped when the user is pinned to bottom. Eviction pauses during scroll-back.
- Tracks latest `agent.status` per agent for the activity indicator
- Exposes a `filteredByAgent` state for the agent name filter

**Auto-scroll logic:** extract a shared `useAutoScroll` hook from the existing `LogStream` component (`web/src/features/agents/LogStream.tsx`), then use it in both `ProjectLogStream` and the original `LogStream`. Avoids duplicating scroll behavior.

Agent color assignment: deterministic hash of agent ID modulo the palette array. Computed client-side via `useAgentColor`.

### New components

| Component | Location | Purpose |
|---|---|---|
| `LogsDrawer` | `web/src/features/logs/LogsDrawer.tsx` | Drawer container with resize handle, title bar, collapse toggle |
| `LogsTitleBar` | `web/src/features/logs/LogsTitleBar.tsx` | Progress text, progress bar fill, activity dot |
| `ProjectLogStream` | `web/src/features/logs/ProjectLogStream.tsx` | Unified log stream with agent colors, expandable rows, auto-scroll |
| `LogEntry` | `web/src/features/logs/LogEntry.tsx` | Single log line with colored agent name, timestamp, expandable detail |
| `useProjectLogs` | `web/src/hooks/useProjectLogs.ts` | WebSocket subscription, event aggregation, agent filtering, status tracking |
| `useAgentColor` | `web/src/hooks/useAgentColor.ts` | Deterministic agent ID to color mapping from fixed palette |
| `useAutoScroll` | `web/src/hooks/useAutoScroll.ts` | Extracted from existing `LogStream` — shared scroll-pinning logic |

### Integration point

`LogsDrawer` is added to `ProjectDashboard.tsx` below the `KanbanBoard` component. The dashboard layout changes to a fixed-height flex column (`h-[calc(100vh-<header>)]`). The task data query moves from `KanbanBoard` to `ProjectDashboard` and is passed as props to both `KanbanBoard` and `LogsTitleBar`.

## What this does not include

- Persisted log history (logs are live stream only, not stored beyond the buffer)
- Log search or text filtering
- Log export
- Per-agent tab view (the agent name click-filter covers this use case)
- Migration from global WebSocket to per-topic ChannelHub subscriptions
- Keyboard shortcuts for drawer toggle
- Dedicated `agent.error` event type (uses payload inspection instead)
