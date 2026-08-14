# Digest Engine

A self-hosted RSS/Atom feed aggregator with intelligent digest scheduling. Built in Go for deployment on Kubernetes/K3s.

## Overview

Digest Engine solves the problem of information overload by:
- Polling RSS feeds continuously (capturing score/engagement data over time)
- Storing items in a local database with historical metrics
- Generating scheduled digests (daily/weekly/monthly) with top content
- Filtering noise and de-duplicating across digest periods

## Architecture

```
┌─────────┐      ┌───────┐      ┌──────────┐
│ Poller  │─────>│SQLite │<─────│ Digester │
│(Hourly) │      │ Store │      │ (Cron)   │
└─────────┘      └───────┘      └──────────┘
                                      │
                                      ▼
                                 ┌────────┐
                                 │ Email  │
                                 │ SMTP   │
                                 └────────┘
```

## Features

### Core
- [x] Multi-source RSS/Atom feed polling
- [x] Score tracking (Reddit karma, HN points, PH upvotes)
- [x] Flexible scheduling (daily/weekly/monthly digests)
- [x] De-duplication across digest periods
- [x] Noise filtering via regex blocklists
- [x] Zero-Fluff & Anti-Clickbait enrichment pipeline (see [ENRICHMENT_DESIGN.md](file:///Users/gordonpn/workspace/digest-engine/ENRICHMENT_DESIGN.md))

### Supported Sources
- Reddit (subreddits via `/top/.rss`, `/new/.rss`)
- Hacker News (via hnrss.org)
- Dev.to (tag-specific feeds)
- Lobsters (topic feeds)
- Hashnode (tag feeds)
- Stack Exchange (hot/newest feeds)
- Product Hunt (daily feed, aggregated monthly)

### Operations
- Prometheus metrics (`/metrics` endpoint)
- Structured logging (JSON)
- Kubernetes-ready (CronJob + Deployment manifests)

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
