# ADR 0001: Northflank with PostgreSQL and a Future Self-Hosted SQLite Profile

- Status: Accepted
- Date: 2026-10-26

## Context

Digest Engine's first deployment target is Northflank with PostgreSQL. The application must also remain practical to self-host on Docker, Docker Swarm, or K3s, where a local SQLite database is preferred. The existing design sketches assume SQLite everywhere, so implementing them unchanged would make the hosted target diverge from the selected deployment.

The product is initially a single-user reading application. It does not need a microservice topology to support either deployment target.

## Decision

- Use Northflank with PostgreSQL for the first hosted deployment.
- Keep the Go application as a modular monolith. The first deployment uses one application service for the reader and background ingestion.
- Keep database-specific drivers, SQL, and migrations in storage adapters. Domain and application behavior must depend on repository contracts rather than PostgreSQL types or dialect-specific behavior.
- Preserve a future SQLite deployment profile for Docker, Docker Swarm, or K3s. The initial hosted release does not require deploying that profile, but its addition must not require changing domain or application contracts.
- Ingest newsletters through a dedicated address from the start. Select an inbound transport that can deliver the message to the application without granting access to the user's general mailbox; record its setup and recovery steps when implemented.

## Consequences

- PostgreSQL provides the hosted database target without making application behavior depend on Northflank APIs.
- SQLite remains an option for a simple, single-instance self-hosted deployment.
- Supporting two database engines requires separate query and migration implementations and verification against each supported engine when the SQLite profile is added.
- SQLite deployments must keep the database on durable local storage and run a single writer instance. Do not share one SQLite file between replicas.
- The inbound email transport is still an implementation decision. It must support the dedicated-address boundary and fit the hosted and self-hosted deployment goals.

## Reconsideration Triggers

- Reconsider the SQLite profile if maintaining equivalent behavior requires substantial dialect-specific logic outside storage adapters.
- Reconsider the single application service if measured ingestion or request load requires independent scaling.
- Reconsider PostgreSQL as the hosted database if the selected Northflank plan cannot provide the required persistence, backup, or recovery guarantees.
