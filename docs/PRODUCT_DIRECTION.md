# Product Direction

## Purpose

Digest Engine is a self-hosted personal information reader. It collects material from feeds and newsletters, helps the reader find useful stories, and keeps those stories available for later reading. The web reader is the primary experience; email delivery is optional.

This document records deferred product ideas. It is a concept inventory, not a second execution backlog. The initial implementation batch is tracked under [GitHub epic #3](https://github.com/gordonpn/digest-engine/issues/3); ready work is tracked in GitHub Issues.

## Accepted Constraints

- The first hosted target is Northflank with PostgreSQL.
- The application remains a Go modular monolith. Keep database drivers, queries, and migrations behind storage adapters so a future SQLite profile can run on Docker, Docker Swarm, or K3s without changing domain behavior.
- The first user-facing MVP slice is newsletter ingestion through a managed inbound-email service and a dedicated hosted address. Deliver this path before RSS/Atom ingestion; the PostgreSQL storage boundary is its prerequisite.
- Keep provider-specific email delivery at the ingress boundary so a future self-hosted profile can use a different transport without changing newsletter processing. Do not connect to or ingest from the user's general personal inbox.
- RSS and Atom are the first feed protocols. Substack and Medium content enters through public RSS feeds when available; do not bypass paywalls or depend on undocumented scraping.
- The reader remains single-user initially. Use one server-side rendering owner; an AHA-style Go HTML and HTMX interface is the preferred direction, while the exact renderer remains an implementation choice.
- Preserve original links and source attribution. Treat fetched pages and newsletter bodies as untrusted content.

## Deferred Feature Candidates

These ideas are intentionally outside the first implementation batch. Promote one to GitHub when its scope and acceptance criteria are ready.

| Candidate | User value | Dependency or trigger |
|---|---|---|
| Full-text searchable archive | Find previously ingested articles and newsletters by source, topic, or date | Stored content and retention policy; start with the database's built-in full-text search |
| Reading notes, highlights, and collections | Keep personal context alongside source material | A working reader and persistent reading state |
| Explicit feedback and explainable ranking | Improve relevance while showing why an item appeared | Baseline ranking and save/read actions |
| Interest profiles and controlled discovery | Follow topics across sources while reserving room for unfamiliar material | Enough explicit feedback to avoid guessing at preferences |
| Story development and corrections | Surface material changes, new evidence, or corrections without repeating old coverage | Reliable cross-source article identity and history |
| ChatGPT exploration handoff | Open a source-aware prompt and add a personal question while reading | Reader page; include a copy-prompt fallback and avoid placing private newsletter bodies in URLs |
| Newsletter source-value report | See which subscriptions contribute relevant, original material versus duplicates or promotions | Source attribution plus read/save feedback; use observable counts rather than subjective fluff percentages |
| Engagement history and trend scoring | Distinguish steady popularity from stories gaining attention quickly | Source-specific engagement data and snapshots; scores from different services need normalization |
| Source health and replay tools | Diagnose stale feeds, parse failures, and failed enrichment | Background ingestion and persisted processing status |
| Multiple briefing and delivery policies | Create daily, weekly, or alert views from the same stored material | Persistent briefing editions and a reader-first baseline |
| Portable export and restore | Preserve the personal archive across hosting changes | Defined data-retention and raw-content policies |
| Additional source discovery | Add sources such as GitHub trending repositories or X | Define the supported query, data access method, limits, and source attribution first; see existing [GitHub trending issue #2](https://github.com/gordonpn/digest-engine/issues/2) and [X issue #1](https://github.com/gordonpn/digest-engine/issues/1) |
| Bounded follow-up research | Investigate important unanswered questions using cited evidence | Trustworthy source retrieval, evaluation criteria, and an explicit cost limit |

## Deferred Platform Complexity

Do not introduce multi-user tenancy, a vector database, an autonomous agent framework, a browser extension, or separately deployed microservices until a measured product need requires them. These are not prerequisites for a useful personal reader.
