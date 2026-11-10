// Package domain contains newsletter models independent of delivery and storage adapters.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"
)

// Source is a configured ingestion origin, initially the dedicated newsletter
// mailbox. Message-level sender attribution remains on each SourceDocument.
type Source struct {
	ID   string
	Name string
}

// SourceDocument retains original source material. All sender-provided fields,
// including BodyText, are untrusted and must be escaped by any renderer.
type SourceDocument struct {
	ID                string
	SourceID          string
	DeliveryKey       string
	Provider          string
	ProviderMessageID string    // Optional; diagnostic provenance, never the deduplication key.
	EnvelopeSender    string    // SMTP MAIL FROM; may be empty for a null reverse path.
	Recipient         string    // Canonical dedicated envelope recipient, never the MIME To header.
	ReceivedAt        time.Time // First authenticated application receipt, in UTC.
	StoredAt          time.Time // First successful durable write, in UTC.
	MessageID         string    // Untrusted MIME Message-ID, possibly missing or reused.
	From              string    // Original MIME From header for reader attribution.
	Subject           string
	BodyText          string // Decoded readable text; no trusted HTML type.
	RawMIME           []byte // Exact provider-received bytes; immutable after first insert.
}

// DocumentSummary is the bounded reader-list projection; it excludes message
// bodies and raw MIME. Source contains the document's source attribution.
type DocumentSummary struct {
	ID         string
	Source     Source
	From       string
	Subject    string
	ReceivedAt time.Time
}

// CanonicalRecipient accepts a bare mailbox, lowercases only the domain, and
// preserves the local part. No plus-address stripping or case folding is done.
func CanonicalRecipient(recipient string) (string, error) {
	address, err := mail.ParseAddress(recipient)
	if err != nil || address.Name != "" || address.Address != recipient {
		return "", errors.New("recipient must be a bare mailbox address")
	}
	at := strings.LastIndexByte(address.Address, '@')
	if at < 1 || at == len(address.Address)-1 || strings.ContainsAny(recipient, "\r\n\x00") {
		return "", errors.New("recipient must be a bare mailbox address")
	}
	return recipient[:at+1] + strings.ToLower(recipient[at+1:]), nil
}

// NewDeliveryKey identifies the same raw message delivered to the same mailbox.
// It deliberately excludes transport tokens, timestamps, sender claims, and MIME
// Message-ID. Any MIME byte change creates a new document, not a semantic merge.
func NewDeliveryKey(recipient string, rawMIME []byte) (string, error) {
	canonical, err := CanonicalRecipient(recipient)
	if err != nil {
		return "", err
	}
	if len(rawMIME) == 0 {
		return "", errors.New("raw MIME is required")
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte(canonical))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write(rawMIME)
	return "mime-v1:" + hex.EncodeToString(digest.Sum(nil)), nil
}
