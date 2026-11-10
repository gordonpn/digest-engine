# Digest Engine

A personal information reader that collects RSS/Atom feeds and newsletters, then helps surface and revisit useful stories.

> **Project status:** The Go HTTP bootstrap is implemented. Storage, ingestion, and reader features are still planned; feature checkboxes below describe that planned behavior.

See [Product Direction](docs/PRODUCT_DIRECTION.md) and [ADR 0001](docs/adr/0001-northflank-postgresql-and-self-hosted-sqlite.md) for current decisions and deferred ideas.

The first user-facing MVP slice is ingesting newsletters through a managed service and a dedicated hosted address. RSS/Atom ingestion follows that path.

## Overview

Digest Engine solves the problem of information overload by:
- Polling RSS feeds and ingesting newsletters sent to a dedicated address
- Storing source material and extracted items with provenance
- Surfacing useful items in a persistent web reader
- Offering scheduled digests as an optional delivery path

## Architecture

```
┌─────────┐      ┌───────┐      ┌──────────┐
│ Sources │─────>│Ingest │─────>│Database │
│ RSS/Email│     │       │      │         │
└─────────┘      └───────┘      └──────────┘
                                      │
                         ┌────────────┴──────────┐
                         ▼                       ▼
                    ┌────────┐             Optional email
                    │Reader  │
                    └────────┘
```

## Features

### Core
- [ ] Multi-source RSS/Atom and newsletter ingestion
- [ ] Source-aware ranking and reading history
- [ ] Cross-source article de-duplication
- [ ] Noise filtering and content-aware extraction
- [ ] Zero-Fluff & Anti-Clickbait enrichment pipeline (see [ENRICHMENT_DESIGN.md](ENRICHMENT_DESIGN.md))

### Planned Sources
- Reddit (subreddits via `/top/.rss`, `/new/.rss`)
- Hacker News (via hnrss.org)
- Dev.to (tag-specific feeds)
- Lobsters (topic feeds)
- Hashnode (tag feeds)
- Stack Exchange (hot/newest feeds)
- Product Hunt (daily feed, aggregated monthly)
- Substack and Medium public RSS feeds
- Newsletters sent to a dedicated address

### Deployment Targets
- First hosted target: Northflank with PostgreSQL
- Self-hosted target: Docker, Docker Swarm, or K3s with SQLite
- One Go application with database-specific behavior isolated behind storage adapters

## Quick Start

### Local Development

Requires Go 1.25 or newer. From the repository root:

```bash
go test -race ./...
go vet ./...
go run ./cmd/digest-engine
```

In another terminal, `curl -fsS http://localhost:8080/livez` returns `ok`.
For a binary that receives signals directly:

```bash
go build -o /tmp/digest-engine ./cmd/digest-engine
PORT=9090 /tmp/digest-engine
```

Send SIGINT (Ctrl-C) or SIGTERM to stop accepting connections and drain active
requests. Shutdown has a bounded grace period; remaining connections are closed
when it expires. A startup failure or shutdown timeout exits with status 1.

| Environment variable | Default | Validation |
|---|---|---|
| `PORT` | `8080` | Integer from 1 through 65535; listens on all interfaces |
| `DIGEST_ENGINE_SHUTDOWN_TIMEOUT` | `10s` | Positive Go duration, such as `15s` |

Configuration errors name the field without logging its value. The bootstrap
uses only these environment variables; it does not load `config.example.yaml`.
`GET /livez` reports process liveness, not database or ingestion readiness.
HTTP header/read/write/idle timeouts are 5/15/30/60 seconds respectively.

### Package layout

- `cmd/digest-engine`: process entry point and signal handling.
- `internal/config`: validated runtime environment configuration.
- `internal/app`: application wiring and HTTP lifecycle.
- `adapters/httpserver`: HTTP routes.
- `domain` and `ports`: reserved for business models and integration contracts.

The existing Dockerfile, YAML configuration, and deployment sketches describe
future functionality. Container deployment, database readiness, `poll`, `digest`,
and `serve` subcommands are not implemented by this bootstrap. The executable
starts the HTTP server directly.

## Configuration

The legacy `config.example.yaml` is a design example, not runtime input yet. Planned sections include:

```yaml
feeds:
  - id: "hn-top"
    url: "https://hnrss.org/newest?points=100"
    schedule: "daily"
    min_score: 100
    blocklist:
      - "(?i)who is hiring"

schedules:
  daily: "0 8 * * *"
  weekly: "0 9 * * 5"
  monthly: "0 9 1 * *"

smtp:
  host: "smtp.gmail.com"
  port: 587
  from: "Digest Engine@yourdomain.com"
  to: "you@example.com"
```

## Project Goals

This project is designed to balance:
1. **Delivery Speed**: Pragmatic library choices where appropriate
2. **Career Growth**: Enterprise patterns (hexagonal architecture, observability)
3. **Fundamentals**: Standard library first, concurrency primitives, raw SQL

See `ARCHITECTURE.md` for design decisions.

## License

MIT
