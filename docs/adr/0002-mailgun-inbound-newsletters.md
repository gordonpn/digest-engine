# ADR 0002: Mailgun Routes for Hosted Newsletter Ingestion

- Status: Accepted for implementation; live provisioning is pending
- Date: 2026-11-09
- Tracking: #18, under #7

## Context

The first user-facing slice receives newsletters at a dedicated address and stores
the original MIME for a web reader. The application runs on Northflank, so the
initial transport must deliver over authenticated HTTPS without access to a
personal mailbox or a self-hosted SMTP service.

## Options considered

The following comparison uses the official documentation linked below. Prices
are public USD list prices reviewed for this decision, not a purchase commitment;
account eligibility, regional terms, taxes, and current quotas must be checked
when provisioning.

| Option | Dedicated address and delivery | Authentication and original MIME | Limits, retries, replay | Cost and operational tradeoff |
|---|---|---|---|---|
| **Mailgun Routes (selected)** | Domain/subdomain MX records; an exact-recipient route forwards directly to the Go HTTPS service | Routes provide HMAC-SHA256 timestamp/token authentication. A URL ending in `mime` provides the `body-mime` original; parsing stays in our application | Application caps MIME at 10 MiB and encoded requests at 32 MiB. HTTP 200 succeeds; 406 permanently rejects; other responses retry for about 8 hours. Optional `store()` retains a recovery copy for up to 3 days, subject to plan retention | Pricing lists Free with 1 inbound route, Basic at $15/month with 5 routes, and Foundation at $35/month with full inbound routing and 1-day message retention. Verify storage entitlement and inbound billing before enabling recovery storage. Direct forwarding needs no AWS bridge |
| **Amazon SES receipt rules + S3 + delivery bridge** | Verified receiving domain and regional MX; receipt rule stores mail in S3; an additional bridge delivers it to Northflank | IAM protects S3 access; raw, unmodified MIME is available. Bridge-to-application authentication must be designed separately | S3 receipt action defaults to a 40 MB maximum. S3 provides a durable replay source; notification/bridge retries and dead-letter handling require additional configuration | SES lists $0.10/1,000 incoming emails plus $0.09/1,000 incoming chunks; S3, notifications, bridge compute and transfer add cost. Low volume cost, but more infrastructure than a direct webhook |

SendGrid Inbound Parse is also viable: its raw mode includes an `email` MIME
field, security policies support payload signatures/OAuth, and its documented
retry horizon is three days. It is a fallback if Mailgun's account terms or
payload stability do not meet the provisioning checks. It is not selected here;
no SendGrid entitlement or price is assumed.

## Decision

Use a Mailgun exact-recipient Route forwarding to
`https://<application-host>/webhooks/mailgun/mime`. The `mime` suffix is required:
Mailgun sends raw MIME in `body-mime` instead of parsed body fields. Keep the
provider adapter isolated from the shared inbound message and repository types.

Use a dedicated receiving subdomain and one configured mailbox. Configure the
route with an anchored, escaped recipient matcher, the HTTPS forwarding action,
and `stop()` so lower-priority routes do not also ingest the message. Do not
redirect the endpoint or rely on MIME To/From headers for routing authorization.

Preserve exact provider-received MIME, not a reconstructed document. Delivery
identity is the versioned recipient-plus-MIME hash agreed with #15; provider
request tokens and MIME Message-ID are not uniqueness keys.

[The ingress contract](../INBOUND_EMAIL.md) defines message fields, authentication,
limits, acknowledgements, time semantics, and recovery boundaries for #19–#23.
No live provider account, DNS changes, subscription, endpoint deployment, or
secret provisioning is performed by this decision.

## Consequences and reconsideration

- Direct HTTPS delivery fits the one-application hosted deployment. A future
  self-hosted ingress produces the same provider-independent message contract.
- The application owns MIME parsing, safe rendering, and durable idempotency.
  Provider authentication does not authenticate the newsletter's claimed author.
- Mailgun form signatures cover timestamp/token, not the complete content.
  TLS is mandatory; signatures and complete payloads must never be logged.
- Optional provider storage gives a short recovery window, not an archive or
  backup. Recovery depends on the actual plan and must be verified before use.
- Exact-MIME hashing deliberately treats a changed MIME document as a new source
  document. If normal retries change MIME bytes, reconsider identity before
  enabling production ingress; do not silently normalize the retained original.
- Reconsider Mailgun if current pricing, receipt limits, region/privacy terms,
  webhook payload stability, or raw-MIME availability fail the provisioning checks.

## Sources

- [Mailgun receiving via HTTP: MIME payload and retry responses](https://documentation.mailgun.com/docs/mailgun/user-manual/receive-forward-store/receive-http)
- [Mailgun Routes webhook signing](https://documentation.mailgun.com/docs/mailgun/user-manual/webhooks/securing-webhooks)
- [Mailgun route actions and temporary storage](https://documentation.mailgun.com/docs/mailgun/user-manual/receive-forward-store/route-actions)
- [Mailgun stored-message retrieval](https://documentation.mailgun.com/docs/mailgun/user-manual/receive-forward-store/storing-and-retrieving-messages)
- [Mailgun pricing](https://www.mailgun.com/pricing)
- [SES S3 receipt action and MIME size limit](https://docs.aws.amazon.com/ses/latest/dg/receiving-email-action-s3.html)
- [SES pricing](https://aws.amazon.com/ses/pricing/)
- [SendGrid raw MIME setup](https://www.twilio.com/docs/sendgrid/for-developers/parsing-email/setting-up-the-inbound-parse-webhook)
- [SendGrid Parse security policies](https://www.twilio.com/docs/sendgrid/for-developers/parsing-email/securing-your-parse-webhooks)
- [SendGrid delivery retries](https://www.twilio.com/docs/sendgrid/for-developers/parsing-email/inbound-email)
