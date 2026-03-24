#!/bin/bash
set -euo pipefail

AGENTS_STATE="/foundry/state/agents.json"

# Initialize state file
if [ ! -f "$AGENTS_STATE" ]; then
    echo '{}' > "$AGENTS_STATE"
fi

register_agent() {
    local role=$1
    local pid=$2
    local worktree=$3

    local tmp=$(mktemp)
    jq --arg role "$role" --arg pid "$pid" --arg worktree "$worktree" \
        '.[$role] = {"pid": $pid, "worktree": $worktree, "status": "active", "started_at": now}' \
        "$AGENTS_STATE" > "$tmp" && mv "$tmp" "$AGENTS_STATE"

    echo "[supervisor] registered agent: $role (PID $pid)"
}

unregister_agent() {
    local role=$1

    local tmp=$(mktemp)
    jq --arg role "$role" '.[$role].status = "stopped"' \
        "$AGENTS_STATE" > "$tmp" && mv "$tmp" "$AGENTS_STATE"

    echo "[supervisor] unregistered agent: $role"
}

launch() {
    local role=$1
    local worktree=$2
    local task_prompt="${3:-Waiting for task assignment via sidecar.}"

    # Read agent definition for model
    local model="sonnet"
    if [ -f "/agents/${role}.md" ]; then
        model=$(grep -m1 "^model:" "/agents/${role}.md" | awk '{print $2}' || echo "sonnet")
    fi

    cd "$worktree"

    # Launch Claude Code agent
    claude --bare \
        -p "$task_prompt" \
        --output-format stream-json \
        --model "$model" \
        --dangerously-skip-permissions \
        --no-session-persistence \
        2>&1 | while IFS= read -r line; do
            echo "[agent:${role}] $line"
        done &

    local agent_pid=$!
    register_agent "$role" "$agent_pid" "$worktree"

    # Monitor the process
    wait $agent_pid 2>/dev/null
    local exit_code=$?

    if [ $exit_code -ne 0 ]; then
        echo "[supervisor] agent $role crashed with exit code $exit_code"

        # Write crash status
        cat > "/shared/status/${role}.json" <<EOF
{
    "role": "$role",
    "status": "crashed",
    "exit_code": $exit_code,
    "timestamp": $(date +%s)
}
EOF
    else
        echo "[supervisor] agent $role completed successfully"
    fi

    unregister_agent "$role"
}

stop_agent() {
    local role=$1
    local pid

    pid=$(jq -r --arg role "$role" '.[$role].pid // empty' "$AGENTS_STATE")
    if [ -z "$pid" ]; then
        echo "[supervisor] agent $role not found"
        return 1
    fi

    echo "[supervisor] stopping agent $role (PID $pid)..."

    # Create pause signal
    touch "/foundry/state/pause-signal"

    # Send SIGTERM, wait up to 30s
    kill "$pid" 2>/dev/null || true
    local waited=0
    while kill -0 "$pid" 2>/dev/null && [ $waited -lt 30 ]; do
        sleep 1
        waited=$((waited + 1))
    done

    # Force kill if still running
    if kill -0 "$pid" 2>/dev/null; then
        echo "[supervisor] force killing agent $role"
        kill -9 "$pid" 2>/dev/null || true
    fi

    rm -f "/foundry/state/pause-signal"
    unregister_agent "$role"
}

# Route commands
case "${1:-}" in
    launch)
        launch "${2:-}" "${3:-}" "${4:-}"
        ;;
    stop)
        stop_agent "${2:-}"
        ;;
    *)
        echo "Usage: supervisor.sh {launch|stop} <role> [worktree] [prompt]"
        exit 1
        ;;
esac
