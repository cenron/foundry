#!/bin/bash
set -euo pipefail

echo "[foundry] team container starting for project $PROJECT_ID"

# Clone the repository
if [ ! -d "/workspace/.git" ]; then
    echo "[foundry] cloning repository..."
    if [ -n "${GIT_TOKEN:-}" ]; then
        git -c "http.extraHeader=Authorization: Bearer ${GIT_TOKEN}" clone "$REPO_URL" /workspace
    else
        git clone "$REPO_URL" /workspace
    fi
else
    echo "[foundry] workspace already cloned, pulling latest..."
    if [ -n "${GIT_TOKEN:-}" ]; then
        cd /workspace && git -c "http.extraHeader=Authorization: Bearer ${GIT_TOKEN}" pull --ff-only || true
    else
        cd /workspace && git pull --ff-only || true
    fi
fi

cd /workspace

# Start heartbeat in background
echo "[foundry] starting heartbeat..."
/foundry/heartbeat.sh &

# Start sidecar in background
echo "[foundry] starting RabbitMQ sidecar..."
node /foundry/sidecar/index.js &
SIDECAR_PID=$!

# Wait for sidecar to connect
sleep 2

# Read team composition
IFS=',' read -ra ROLES <<< "${TEAM_COMPOSITION:-}"

echo "[foundry] team composition: ${ROLES[*]}"

# Create worktrees and launch agents for each role
for ROLE in "${ROLES[@]}"; do
    ROLE=$(echo "$ROLE" | xargs) # trim whitespace
    BRANCH="agent/${ROLE}"
    WORKTREE="/worktrees/${ROLE}"

    if [ ! -d "$WORKTREE" ]; then
        echo "[foundry] creating worktree for $ROLE on branch $BRANCH..."
        git worktree add "$WORKTREE" -b "$BRANCH" 2>/dev/null || \
            git worktree add "$WORKTREE" "$BRANCH" 2>/dev/null || \
            git worktree add "$WORKTREE" -b "$BRANCH" HEAD
    fi

    echo "[foundry] launching agent: $ROLE"
    /foundry/supervisor.sh launch "$ROLE" "$WORKTREE" &
done

echo "[foundry] all agents launched, supervisor monitoring..."

# Wait for sidecar (main process)
wait $SIDECAR_PID
