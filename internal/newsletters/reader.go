// Package newsletters implements application behavior through repository ports.
package newsletters

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gordonpn/digest-engine/domain"
	"github.com/gordonpn/digest-engine/ports"
)

var ErrInvalidQuery = errors.New("invalid newsletter reader query")

type Reader struct{ repository ports.NewsletterRepository }

func NewReader(repository ports.NewsletterRepository) *Reader {
	return &Reader{repository: repository}
}

func (r *Reader) List(ctx context.Context, options ports.ListOptions) ([]domain.DocumentSummary, error) {
	if options.Limit == 0 {
		options.Limit = 20
	}
	if options.Limit < 1 || options.Limit > 100 || options.Offset < 0 {
		return nil, ErrInvalidQuery
	}
	summaries, err := r.repository.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("list newsletters: %w", err)
	}
	return summaries, nil
}

// DocumentView excludes retained MIME and transport metadata. BodyText is still
// untrusted; an HTTP adapter must use contextual escaping rather than trusted HTML.
type DocumentView struct {
	ID       string
	SourceID string
	From     string
	Subject  string
	BodyText string
}

func (r *Reader) Get(ctx context.Context, id string) (DocumentView, error) {
	if strings.TrimSpace(id) == "" {
		return DocumentView{}, ErrInvalidQuery
	}
	doc, err := r.repository.Get(ctx, id)
	if err != nil {
		return DocumentView{}, fmt.Errorf("get newsletter: %w", err)
	}
	return DocumentView{ID: doc.ID, SourceID: doc.SourceID, From: doc.From, Subject: doc.Subject, BodyText: doc.BodyText}, nil
}
