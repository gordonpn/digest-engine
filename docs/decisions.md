# Implementation Decisions Log

This document records non-obvious engineering and implementation decisions made in Digest Engine, ordered newest first. Major architectural decisions that are expensive to reverse are promoted to formal Architecture Decision Records under `docs/adr/`.

## 2026-11-10: Select Mailgun for inbound newsletter webhook ingress

- Decision: Use Mailgun HTTP webhooks to receive inbound newsletter deliveries at a dedicated domain address, verifying webhook HMAC signatures before processing.
- Rationale: Mailgun eliminates managing SMTP inbound mail server daemons or polling IMAP inboxes, providing direct HTTP POST delivery with MIME parsing and authenticity verification.
- Rejected alternatives: Self-hosted Postfix/Haraka SMTP listener, direct IMAP polling against a personal email mailbox.
- Source: [#18 Select the inbound-email provider and define the message contract](https://github.com/gordonpn/digest-engine/issues/18), [#26 Select Mailgun and define the inbound newsletter contract](https://github.com/gordonpn/digest-engine/pull/26)

## 2026-11-10: Bootstrap Go modular monolith with standard library HTTP server

- Decision: Structure the application as a Go modular monolith using standard library `net/http`, hexagonal ports and adapters (`cmd`, `domain`, `ports`, `adapters`, `internal`), and signal-driven graceful shutdown.
- Rationale: A single deployable binary satisfies the single-user personal reader requirements with minimal operational overhead and zero external web framework dependencies.
- Rejected alternatives: Microservices split between ingestion and web reader, heavy third-party Go web frameworks.
- Source: [#14 Bootstrap the Go application](https://github.com/gordonpn/digest-engine/issues/14), [#24 Bootstrap the Go HTTP application](https://github.com/gordonpn/digest-engine/pull/24)

## 2026-11-03: Prioritize hosted newsletter ingestion MVP slice

- Decision: Prioritize hosted newsletter ingestion as the initial end-to-end MVP milestone before implementing RSS/Atom crawlers or public web scraping.
- Rationale: Dedicated email newsletter ingestion addresses the highest friction intake workflow and validates persistence and rendering on the hosted environment first.
- Rejected alternatives: Implementing broad multi-source ingestion (RSS, Substack, GitHub, social aggregations) before completing the reading pipeline.
- Source: [#7 Ingest newsletters through a hosted dedicated address](https://github.com/gordonpn/digest-engine/issues/7), [#12 Prioritize hosted newsletter ingestion MVP slice](https://github.com/gordonpn/digest-engine/pull/12)

## 2026-10-26: Northflank with PostgreSQL and future self-hosted SQLite profile

- Decision: Target Northflank with PostgreSQL for the initial hosted deployment while keeping database drivers, SQL dialects, and migrations isolated behind storage adapters to permit a future self-hosted single-instance SQLite profile.
- Rationale: Hosted Northflank requires PostgreSQL for persistence, but self-hosted single-instance deployments benefit from a zero-configuration local SQLite database without altering core domain logic.
- Rejected alternatives: Hardcoding PostgreSQL types directly in domain entities, requiring an external PostgreSQL instance for self-hosting.
- Source: [ADR 0001](adr/0001-northflank-postgresql-and-self-hosted-sqlite.md), [#4 Keep PostgreSQL and SQLite behind the storage boundary](https://github.com/gordonpn/digest-engine/issues/4)
