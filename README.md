# Digest Engine

A personal information reader that collects RSS/Atom feeds and newsletters, then helps surface and revisit useful stories.

> **Project status:** This repository is a design scaffold with examples; it does not yet contain a Go application. Feature checkboxes below describe planned behavior.

See [Product Direction](docs/PRODUCT_DIRECTION.md) and [ADR 0001](docs/adr/0001-northflank-postgresql-and-self-hosted-sqlite.md) for current decisions and deferred ideas.

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

```bash
# 1. Clone and configure
git clone https://github.com/yourusername/Digest Engine
cd Digest Engine
cp config.example.yaml config.yaml

# 2. Edit config.yaml with your feeds and SMTP settings

# 3. Run locally
go run cmd/Digest Engine/main.go poll   # Manual poll
go run cmd/Digest Engine/main.go digest # Generate digest

# 4. Run with scheduler
go run cmd/Digest Engine/main.go serve  # Starts HTTP + Cron
```

### Docker

```bash
docker build -t Digest Engine:latest .
docker run -v $(pwd)/config.yaml:/app/config.yaml \
           -v $(pwd)/data:/app/data \
           Digest Engine:latest serve
```

### Kubernetes (K3s)

```bash
kubectl apply -f k8s/
```

## Configuration

See `config.example.yaml` for full documentation. Key sections:

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
