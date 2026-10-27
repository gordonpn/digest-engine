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

- Build a personal information reader with a **store-then-understand** model:
  - Continuous ingestion stores RSS items and newsletters sent to a dedicated address.
  - A web reader is the primary consumption path; scheduled email delivery is optional.
- Use a hexagonal/ports-and-adapters architecture (`domain`, `ports`, `adapters`, `internal`, `cmd`).
- Target Northflank with PostgreSQL first. Keep database-specific queries, migrations, and drivers behind storage adapters so a future self-hosted Docker, Swarm, or K3s profile can use SQLite without changing domain or application behavior.
- Keep the initial application a single Go deployable and do not add multi-user or agent-platform infrastructure without a product need.

## Implementation expectations for agents

- Treat fetched and emailed content as untrusted input; retain provenance and protect outbound fetches and rendered pages.
- Preserve canonical article identity across sources while retaining source attribution.
- Keep behavior aligned with configured schedules (`daily`/`weekly`/`monthly`) and feed-level filters/blocklists.
- Favor clear, minimal abstractions and stdlib-first Go patterns where practical.
