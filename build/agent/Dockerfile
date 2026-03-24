FROM ubuntu:24.04

# Prevent interactive prompts during install
ENV DEBIAN_FRONTEND=noninteractive

# System dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    jq \
    openssh-client \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Node.js 22 (for sidecar + Claude Code CLI)
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Go toolchain
RUN curl -fsSL https://go.dev/dl/go1.26.1.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

# Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Runtime directories
RUN mkdir -p /workspace /worktrees /shared /agents \
    /foundry/state /foundry/sidecar

# Foundry runtime scripts
COPY entrypoint.sh supervisor.sh heartbeat.sh /foundry/
RUN chmod +x /foundry/*.sh

# Sidecar (Node.js RabbitMQ bridge)
COPY sidecar.js /foundry/sidecar/index.js
COPY sidecar-package.json /foundry/sidecar/package.json
WORKDIR /foundry/sidecar
RUN npm install --production

# Workspace template (baked into the image)
COPY workspace/ /foundry/workspace-template/

WORKDIR /workspace
ENTRYPOINT ["/foundry/entrypoint.sh"]
