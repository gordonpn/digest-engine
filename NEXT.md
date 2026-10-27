# Next

## NOW

[#4 Keep PostgreSQL and SQLite behind the storage boundary](https://github.com/gordonpn/digest-engine/issues/4)

- Priority: P1
- Effort: M
- Readiness: Ready
- Reason: Establishes the hosted database path and preserves a future SQLite self-host profile before application persistence spreads.
- First action: Create the Go application skeleton and PostgreSQL storage adapter behind repository contracts.
- Definition of Done: Northflank configuration selects PostgreSQL; migrations and integration checks run; domain and application packages do not depend on database drivers; the SQLite extension point is documented.
- Unblocks: [#5 RSS-to-reader](https://github.com/gordonpn/digest-engine/issues/5) and [#7 dedicated-address newsletter ingestion](https://github.com/gordonpn/digest-engine/issues/7).

## NEXT

- [#5 Ingest RSS and Atom items into a web reader](https://github.com/gordonpn/digest-engine/issues/5), P1 / M; depends on #4.
- [#7 Ingest newsletters sent to a dedicated address](https://github.com/gordonpn/digest-engine/issues/7), P1 / M; depends on #4. The transport choice is part of this issue.
- [#6 Add Substack and Medium sources through public RSS](https://github.com/gordonpn/digest-engine/issues/6), P1 / S; depends on #5.
- [#8 Protect article fetching and reader rendering](https://github.com/gordonpn/digest-engine/issues/8), P1 / M; can proceed with #5 after #4.
- #5 and #7 can be implemented in parallel after the storage boundary is ready.

## LATER

- [#9 Preserve roundup stories and long-form essays](https://github.com/gordonpn/digest-engine/issues/9), P1 / M; depends on #7 and #8.
- [#10 Deduplicate identical articles across sources](https://github.com/gordonpn/digest-engine/issues/10), P1 / M; depends on #5 and #7.
- [#11 Add a persistent save and read queue](https://github.com/gordonpn/digest-engine/issues/11), P1 / S; depends on #5.
- Other candidates are recorded in [Product Direction](docs/PRODUCT_DIRECTION.md).

## BLOCKERS

- None for #4. The inbound newsletter transport is intentionally selected within #7.
