// Package ports defines provider-independent application integration contracts.
package ports

import (
	"context"
	"errors"
	"time"

	"github.com/gordonpn/digest-engine/domain"
)

// InboundMessage is produced only after ingress authentication, recipient and
// size validation. It contains no provider payload, HTTP types, or credentials.
// The configured source is selected outside untrusted MIME headers.
type InboundMessage struct {
	SourceID          string
	Provider          string
	ProviderMessageID string
	EnvelopeSender    string
	Recipient         string
	ReceivedAt        time.Time
	RawMIME           []byte
}

var ErrNotFound = errors.New("newsletter document not found")

type ListOptions struct {
	Limit  int // 1..100; application defaults an omitted value to 20.
	Offset int // Nonnegative; single-user MVP pagination.
}

type StoreResult struct {
	DocumentID string // Existing document ID for a duplicate.
	Created    bool
}

// NewsletterRepository is the storage boundary for the newsletter slice.
// Implementations honor context cancellation and wrap errors for errors.Is.
// Driver-specific types, SQL and migrations belong exclusively in adapters.
type NewsletterRepository interface {
	// Store atomically stores source attribution and the complete document.
	// doc.SourceID must equal source.ID. ID, DeliveryKey and canonical Recipient
	// are required; the application computes the key with domain.NewDeliveryKey.
	// A globally unique DeliveryKey makes concurrent retries insert exactly once.
	// A duplicate returns its original ID with Created=false and never overwrites
	// source, raw MIME, normalized content, ReceivedAt or StoredAt.
	// No successful result is returned before the transaction commits.
	Store(ctx context.Context, source domain.Source, doc domain.SourceDocument) (StoreResult, error)
	// Get returns the complete retained source document or ErrNotFound.
	// Returned mutable byte slices must not alias adapter-owned state.
	Get(ctx context.Context, id string) (domain.SourceDocument, error)
	// List returns summaries joined to their source, ordered ReceivedAt DESC,
	// then ID DESC to break ties. Empty results are successful; invalid options
	// are rejected. It must not load raw MIME or message bodies for listing.
	List(ctx context.Context, options ListOptions) ([]domain.DocumentSummary, error)
}
