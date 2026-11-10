# Architecture Decision Records

| ID | Title | Status | Date | Decision |
|---|---|---|---|---|
| [0001](0001-northflank-postgresql-and-self-hosted-sqlite.md) | Northflank with PostgreSQL and a future self-hosted SQLite profile | Accepted | 2026-10-26 | Use Northflank and PostgreSQL for the first hosted target; keep application behavior independent of database adapters so self-hosting can use SQLite. |
| [0002](0002-mailgun-inbound-newsletters.md) | Mailgun Routes for hosted newsletter ingestion | Accepted for implementation | 2026-11-09 | Receive raw MIME over authenticated HTTPS at a dedicated address; retain provider-independent message and storage contracts. |
