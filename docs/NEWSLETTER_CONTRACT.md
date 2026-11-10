# Newsletter domain and storage contract

This is the shared contract for #15 and #18. The Go definitions in `domain` and
`ports` are authoritative for types. No storage adapter or ingestion endpoint is
implemented here; #16 and #21 implement the persistence guarantees below.

## Ownership and flow

1. A transport adapter authenticates delivery, enforces byte limits, verifies the
   envelope recipient against the configured dedicated mailbox, and selects a
   configured source ID. It creates `ports.InboundMessage`, retaining exact MIME
   bytes. Provider payloads and credentials stay at the adapter boundary.
2. The MIME parser (#19) produces From, Subject, Message-ID and readable plain
   text without modifying the retained original. Envelope sender and MIME From
   are distinct provenance fields. Neither proves the author's identity.
3. The application (#21) constructs a `domain.SourceDocument`, computes its
   delivery key, assigns an opaque ID, and invokes `NewsletterRepository.Store`.
4. The PostgreSQL adapter (#16) atomically inserts the source and complete
   document. A database uniqueness constraint on DeliveryKey arbitrates retries,
   including simultaneous deliveries. No check-then-insert implementation.
5. Reader application behavior consumes the repository without database or
   provider types. A future SQLite adapter implements the same contract.

A Source initially represents the configured dedicated mailbox, not a discovered
publisher or the untrusted MIME From value. Documents retain their individual
sender attribution. Automatic publisher discovery is outside this slice.

## Identity and time

The versioned delivery key is `mime-v1:` followed by lowercase hexadecimal
SHA-256 of `canonical envelope recipient || NUL || exact raw MIME bytes`.
Canonicalization lowercases only the domain of a bare mailbox. Preserve local
part case and plus tags; compare incoming recipients to the configured canonical
mailbox before constructing a message. Display names and multiple recipients
are not accepted as the envelope recipient.

This deduplicates identical webhook retries and exact saved-MIME replays, including
across transports. Provider webhook tokens, authentication timestamps and optional
provider message IDs are not stable identity. MIME Message-ID may be absent,
forged or reused and is retained only as provenance. Distinct bytes create distinct
documents, even with the same title or Message-ID. Retransmission that modifies
Received headers is a new document; semantic/cross-source merging is #10.

ReceivedAt is UTC application time at the first authenticated ingress attempt
that is eventually persisted; it is not MIME Date or the webhook signing time.
StoredAt is UTC at first durable insertion. A failed attempt has no durable time;
a retry may supply a later ReceivedAt. A duplicate never changes either stored
timestamp, the original ID, source attribution, or content.

## Persistence and retrieval invariants

- Source and document IDs are nonempty opaque application identifiers.
- The supplied source ID equals the document SourceID; the key is recomputed
  from the canonical recipient and retained MIME before persistence.
- Raw MIME is required and immutable after insertion. Store it together with
  normalized content; acknowledge only after commit. Failed transactions leave
  no partial source/document writes.
- A successful duplicate returns the original DocumentID and Created=false.
  A new insert returns Created=true. Never report success for an uncommitted write.
- Get returns a complete document or `ports.ErrNotFound`. Returned byte slices
  must not expose mutable adapter-owned state.
- List reads only summaries, ordered ReceivedAt descending then ID descending.
  Source attribution is joined into each summary. Limits are 1 through 100,
  offsets are nonnegative; the application defaults an omitted limit to 20.
  Empty results succeed. Offset paging can shift under concurrent ingestion;
  snapshot-stable pagination is outside the single-user MVP.
- Honor context cancellation and preserve error identity for `errors.Is`.

## Content handling

Keep the raw original for auditing and reprocessing; do not log it, return it in
reader list projections, or convert it to trusted HTML. `DocumentView` excludes
raw MIME and transport metadata. BodyText remains untrusted even after MIME
normalization; future renderers must use contextual escaping. HTML-to-text
conversion, active-content removal, resource bounds, and HTTP rendering are
implemented in #19, #22 and #8, not by these model types.

## Verification

Run `go test -race ./...` and `go vet ./...`. Tests lock down the identity
algorithm, distinguish changed content/mailboxes from retries, reject invalid
identity inputs, and exercise reader behavior through repository test doubles.
Actual transaction, uniqueness, and concurrent retry guarantees require the
PostgreSQL integration checks in #16 and #21; test doubles do not establish them.
