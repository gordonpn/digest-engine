# AGENTS.md

Guidance for coding agents working in this repository.

## Do not rely on this file alone

Always read **all markdown files** in the repo before implementing changes.  
At minimum, review these documents alongside `AGENTS.md`:

- `README.md` — product overview, features, local run flow, and config examples.
- `ARCHITECTURE.md` — design principles, system architecture, data model, concurrency, observability, and deployment trade-offs.
- `IMPLEMENTATION.md` — phased implementation blueprint and example code structure.

If new `.md` files are added later, treat them as part of the source guidance too.

## Project intent (from existing docs)

- Build a self-hosted feed digest engine with a **store-then-digest** model:
  - Continuous polling stores and updates items.
  - Scheduled digest jobs rank and send top items.
- Use a hexagonal/ports-and-adapters architecture (`domain`, `ports`, `adapters`, `internal`, `cmd`).
- Prioritize observability (structured logs + Prometheus metrics) and operational simplicity (SQLite-first, K8s-ready).

## Implementation expectations for agents

- Preserve de-duplication across digest periods (via digest history tracking).
- Keep behavior aligned with configured schedules (`daily`/`weekly`/`monthly`) and feed-level filters/blocklists.
- Favor clear, minimal abstractions and stdlib-first Go patterns where practical.
