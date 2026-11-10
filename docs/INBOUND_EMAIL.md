# Hosted inbound-email contract

Issue #18 defines this transport boundary; companion issue #15 / PR #25 defines
the Go `ports.InboundMessage`, newsletter domain models, and storage contracts.
This document is an implementation contract for #19–#23, not an implemented
endpoint or an operationally verified provider account.

## Provider-independent message

An ingress adapter emits one message only after authentication, bounded request
parsing, and envelope-recipient validation. Provider-specific HTTP fields never
reach domain or application packages.

| Field | Meaning and Mailgun mapping |
|---|---|
| SourceID | Configured dedicated-mailbox source ID; never derived from untrusted MIME headers |
| Provider | Adapter identifier `mailgun`; provenance only |
| ProviderMessageID | Optional stable provider identifier if actually available; empty for the selected raw-MIME payload, which guarantees no such field |
| EnvelopeSender | Mailgun `sender` (SMTP MAIL FROM), which may be empty; distinct from MIME From |
| Recipient | Mailgun `recipient` (SMTP envelope), matched against the configured canonical dedicated mailbox |
| ReceivedAt | UTC application time captured after ingress authentication; not MIME Date or webhook signing time |
| RawMIME | Exact decoded bytes from `body-mime`, without line-ending, whitespace, or header normalization |

The parser extracts MIME From, Subject, Message-ID and readable plain text while
retaining RawMIME unchanged. No enrichment, URL fetching, attachment execution,
or roundup extraction occurs at this boundary. Store original content and
source attribution even when normalized content will be reprocessed later.

A Source initially identifies the configured dedicated mailbox; each document
retains its individual sender provenance. Authentication proves delivery via
the provider, not that the MIME author identity is trustworthy.

## Stable identity and received time

Use the algorithm implemented by #15:

```text
canonical recipient = bare mailbox with domain lowercased; local part unchanged
DeliveryKey = "mime-v1:" + hex(SHA256(canonical recipient + NUL + RawMIME))
```

Preserve local-part case and plus tags. Both the configured mailbox and the
incoming envelope address use the same canonicalization, then must match
exactly. MIME To may be absent, different, or contain multiple addresses; it is
not the authorization input.

The application assigns an opaque document ID. The storage adapter enforces a
unique DeliveryKey with an atomic transaction. Concurrent duplicate deliveries
return the existing document ID and are successful without overwriting the
original content, ID, source, ReceivedAt or StoredAt.

ReceivedAt records the authenticated attempt that is first successfully persisted.
An earlier failed attempt has no durable receipt time. StoredAt is UTC at that
first insertion. Retry/replay times never replace existing timestamps.

Webhook tokens, timestamps, optional provider IDs, and MIME Message-ID are not
delivery identity. Identical saved-MIME replay to the same mailbox is a duplicate;
modified MIME (including changed Received headers) is a new document. Semantic
and cross-source deduplication remains #10.

## Mailgun ingress requirements

- Accept HTTPS POST at `/webhooks/mailgun/mime`, using the documented raw-MIME
  `application/x-www-form-urlencoded` payload. Reject unsupported media types;
  reject duplicate authentication, recipient, sender, or body-mime fields rather
  than choosing an ambiguous value. Require a nonempty body-mime field.
- Bound the request before parsing: default maximum encoded request is 32 MiB,
  decoded MIME is 10 MiB, and aggregate non-MIME field content is 64 KiB. Percent
  encoding can expand MIME to three times its byte length. Never use an unbounded
  `ReadAll` or trust Content-Length. Enforce the decoded limits too.
- These are application limits, not claims about provider account limits.
  Provisioning must verify that the provider and Northflank permit the selected
  request size and timeout; use the smaller effective limit. Keep parsing and
  attachment expansion bounded separately in #19.
- Require top-level timestamp, token and signature fields. Compute
  HMAC-SHA256(signing key, timestamp || token) and compare decoded signatures in
  constant time. Use the account HTTP webhook signing key, not the API key.
- Accept a signing timestamp no more than 9 hours old or 5 minutes in the future.
  The age window accommodates Mailgun's documented roughly 8-hour retry horizon;
  never use a strict 5-minute expiry that would reject legitimate delayed retries.
  Confirm retry signature behavior with provider fixtures/live checks in #23.
- Validate signatures on every attempt, including duplicates. The form signature
  covers timestamp/token, not the MIME or recipient; require TLS and protect
  tokens as credentials. Never treat a provider signature as author verification.
- Do not reject a legitimate retry merely because its token was previously seen.
  Durable recipient-plus-MIME identity provides insert-once behavior; no in-memory
  cache may claim durable success before the repository has committed.
- Match the canonical envelope recipient to the one configured dedicated mailbox.
  Do not accept catch-all addresses, consult a personal mailbox, or fetch arbitrary
  URLs supplied in provider metadata.
- Map into the shared message only after the checks pass. Decode MIME locally;
  use no provider-parsed or stripped body as a replacement for the retained original.

## Acknowledgement and failure handling

A 200 response is permitted only after the complete raw and normalized document
has committed, or storage confirms an already committed duplicate. There is no
in-memory queue acknowledged before persistence in this slice.

| Condition | Response | Persistence and recovery |
|---|---|---|
| New document committed or committed duplicate | 200 | Provider stops retrying |
| Missing/invalid signature or expired timestamp | 401 | No storage; Mailgun retries, allowing configuration repair within its retry window |
| Wrong recipient, unsupported media, malformed envelope/MIME, or application size limit exceeded | 406 | No storage; Mailgun permanently stops retrying. Return a static rejection reason and record a content-free diagnostic |
| Parser internal failure, database unavailable, or uncertain transaction result | 503 | No success acknowledgement; retry is safe because delivery identity is durable |

Mailgun's documented special cases are 200 and 406; other statuses retry for
about eight hours (10m, 15m, 30m, 1h, 2h, 4h intervals). Do not use 202 for a
non-durable queue or assume a generic 4xx stops retries. Infrastructure may emit
413 before the application; that is a retryable failure from Mailgun's perspective.

Report counts and static failure codes for rejected and failed deliveries. Do not
log MIME, subjects, sender/recipient addresses, signatures, keys, or complete
provider payloads. Authentication/size rejection never stores a source document.

## Configuration prerequisites

These configuration names are proposed for the ingress implementation, not
variables consumed by the HTTP bootstrap today:

| Setting | Purpose |
|---|---|
| `INBOUND_SOURCE_ID` | Stable configured mailbox source identifier |
| `INBOUND_RECIPIENT` | One bare dedicated mailbox, canonicalized and validated at startup |
| `MAILGUN_WEBHOOK_SIGNING_KEY` | Secret signing key injected through hosted secret configuration |
| `INBOUND_MAX_REQUEST_BYTES` | Positive encoded limit, default 33554432 |
| `INBOUND_MAX_MIME_BYTES` | Positive decoded MIME limit, default 10485760 |

Provisioning needs ownership of a receiving subdomain, its Mailgun account/region,
MX records, an exact-recipient route, a public Northflank HTTPS URL, and a webhook
signing key. Verify current plan eligibility, inbound quotas, payload sizes,
privacy/retention terms, and retry behavior before subscribing real newsletters.
No general-mailbox credentials are required. An API key is unnecessary for normal
forwarding; retrieval of optional provider-stored messages uses a separate
least-privilege operational credential, never a webhook key.

## Replay and verification plan

1. For automatic retries, repair provider/app configuration before the eight-hour
   window expires. Repeated valid delivery must return the original document ID.
2. Optionally add `store()` to the exact-recipient route before forwarding and
   `stop()`. Verify the plan's actual retention first: route documentation says up
   to three days, while Foundation pricing currently lists one-day message retention.
   Do not assume three days are available for this account.
3. Retrieve stored MIME through the authenticated provider Messages API with its
   documented raw-MIME Accept mode (`message/rfc2822`). Use a trusted operator path;
   never let arbitrary inbound URLs trigger a fetch. Preserve the original bytes.
4. In #23, provide a local/operator replay path that submits those exact bytes and
   the original recipient through the same application contract with fresh,
   authenticated operator provenance. Do not forge or reuse expired webhook
   signatures. A missing document receives a new application receipt time; an
   existing document keeps its original timestamps.
5. After provider retention expires, replay requires a separately saved original
   or sender resend. Successfully stored originals can be reprocessed locally.
   Changed MIME may create a new document; this is an explicit identity limit.

#23 must verify raw-MIME extraction, valid/invalid signatures, delayed retries,
wrong recipients, percent-encoding expansion, permanent rejection, transient
storage failure, concurrent duplicates, reader visibility, and replay. Document
fixture-only checks separately from live-provider checks and unmet provisioning
prerequisites. Full executable setup/replay instructions belong to #23 once the
endpoint and replay tooling exist; this decision creates neither.
