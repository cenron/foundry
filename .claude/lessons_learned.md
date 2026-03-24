# Lessons Learned

<!-- Entry format:
## [YYYY-MM-DD] Short title
**What happened:** Description
**Takeaway:** The rule or insight
-->

## [2026-03-23] Postgres port conflict with soapbox project
**What happened:** Docker Compose tried to bind Postgres on port 5432 but another project (soapbox) already had a container on that port.
**Takeaway:** Foundry uses port 5433 for Postgres to avoid conflicts. The DATABASE_URL default reflects this.

## [2026-03-23] Use rabbitmq/amqp091-go, not streadway/amqp
**What happened:** CLAUDE.md references `streadway/amqp` but that library is deprecated and archived.
**Takeaway:** Use `github.com/rabbitmq/amqp091-go` — it's the maintained fork with the same API. Import alias: `amqp "github.com/rabbitmq/amqp091-go"`.

## [2026-03-23] golangci-lint not on PATH by default
**What happened:** `golangci-lint` was installed to `~/go/bin/` but that directory wasn't on PATH in the shell.
**Takeaway:** Run with `export PATH="$HOME/go/bin:$PATH"` before `golangci-lint run`, or use the full path.
