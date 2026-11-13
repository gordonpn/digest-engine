# Next

## NOW

[#4 Keep PostgreSQL and SQLite behind the storage boundary](https://github.com/gordonpn/digest-engine/issues/4)

- Priority: P1
- Effort: M
- Readiness: Ready
- Reason: This is the minimum storage foundation required before the first user-facing MVP slice: hosted newsletter ingestion in #7.
- First action: Create the Go application skeleton and PostgreSQL storage adapter behind repository contracts.
- Definition of Done: Northflank configuration selects PostgreSQL; migrations and integration checks run; domain and application packages do not depend on database drivers; the SQLite extension point is documented.
- Unblocks: [#7 hosted newsletter ingestion](https://github.com/gordonpn/digest-engine/issues/7), the first user-facing slice, then [#5 RSS-to-reader](https://github.com/gordonpn/digest-engine/issues/5).

## NEXT

- [#7 Ingest newsletters through a hosted dedicated address](https://github.com/gordonpn/digest-engine/issues/7), P1 / M; depends on #4 and is the first user-facing MVP slice. Mailgun Routes is selected in [PR #26](https://github.com/gordonpn/digest-engine/pull/26) for #18; account/DNS provisioning, endpoint implementation, and live delivery/replay verification remain pending under #20 and #23.
- [#5 Ingest RSS and Atom items into a web reader](https://github.com/gordonpn/digest-engine/issues/5), P1 / M; follows #7 in MVP delivery order and depends on #4.
- [#6 Add Substack and Medium sources through public RSS](https://github.com/gordonpn/digest-engine/issues/6), P1 / S; depends on #5.
- [#8 Protect article fetching and reader rendering](https://github.com/gordonpn/digest-engine/issues/8), P1 / M; can proceed after #4.

## LATER

- [#9 Preserve roundup stories and long-form essays](https://github.com/gordonpn/digest-engine/issues/9), P1 / M; depends on #7 and #8.
- [#10 Deduplicate identical articles across sources](https://github.com/gordonpn/digest-engine/issues/10), P1 / M; depends on #5 and #7.
- [#11 Add a persistent save and read queue](https://github.com/gordonpn/digest-engine/issues/11), P1 / S; depends on #5.
- Other candidates are recorded in [Product Direction](docs/PRODUCT_DIRECTION.md).

## BLOCKERS

- None for #4. Mailgun Routes is selected for #7; hosted ingestion still requires account/DNS provisioning and endpoint implementation (#20), followed by live delivery/replay verification (#23).
