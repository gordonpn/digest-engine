package newsletters

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gordonpn/digest-engine/domain"
	"github.com/gordonpn/digest-engine/ports"
)

type repositoryStub struct {
	list func(context.Context, ports.ListOptions) ([]domain.DocumentSummary, error)
	get  func(context.Context, string) (domain.SourceDocument, error)
}

var _ ports.NewsletterRepository = (*repositoryStub)(nil)

func (s *repositoryStub) Store(context.Context, domain.Source, domain.SourceDocument) (ports.StoreResult, error) {
	panic("unexpected write from reader")
}
func (s *repositoryStub) List(ctx context.Context, o ports.ListOptions) ([]domain.DocumentSummary, error) {
	return s.list(ctx, o)
}
func (s *repositoryStub) Get(ctx context.Context, id string) (domain.SourceDocument, error) {
	return s.get(ctx, id)
}

func TestListBoundsAndAttribution(t *testing.T) {
	expected := []domain.DocumentSummary{{ID: "doc-1", Source: domain.Source{ID: "mailbox", Name: "Newsletters"}, From: "Author <author@example.com>", Subject: "Essay"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	called := false
	repo := &repositoryStub{list: func(got context.Context, o ports.ListOptions) ([]domain.DocumentSummary, error) {
		called = true
		if got != ctx || o.Limit != 20 || o.Offset != 0 {
			t.Fatalf("unexpected repository query: %+v", o)
		}
		return expected, nil
	}}
	reader := NewReader(repo)
	actual, err := reader.List(ctx, ports.ListOptions{})
	if err != nil || !reflect.DeepEqual(actual, expected) {
		t.Fatalf("source attribution lost: %+v, %v", actual, err)
	}
	if !called {
		t.Fatal("repository was not called")
	}
	repo.list = func(context.Context, ports.ListOptions) ([]domain.DocumentSummary, error) {
		t.Fatal("invalid query reached storage")
		return nil, nil
	}
	for _, o := range []ports.ListOptions{{Limit: -1}, {Limit: 101}, {Offset: -1}} {
		if _, err := reader.List(ctx, o); !errors.Is(err, ErrInvalidQuery) {
			t.Fatalf("expected validation error, got %v", err)
		}
	}
}

func TestGetReaderProjectionAndMissingDocument(t *testing.T) {
	ctx := context.Background()
	repo := &repositoryStub{get: func(got context.Context, id string) (domain.SourceDocument, error) {
		if got != ctx || id != "doc-1" {
			t.Fatal("incorrect lookup")
		}
		return domain.SourceDocument{ID: id, SourceID: "mailbox", From: "Author", Subject: "Essay", BodyText: "<script>untrusted</script>", ReceivedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), RawMIME: []byte("private original"), Provider: "provider"}, nil
	}}
	reader := NewReader(repo)
	view, err := reader.Get(ctx, "doc-1")
	if err != nil {
		t.Fatal(err)
	}
	expected := DocumentView{ID: "doc-1", SourceID: "mailbox", From: "Author", Subject: "Essay", BodyText: "<script>untrusted</script>", ReceivedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	if view != expected {
		t.Fatalf("unexpected reader projection: %+v", view)
	}
	repo.get = func(context.Context, string) (domain.SourceDocument, error) {
		return domain.SourceDocument{}, ports.ErrNotFound
	}
	if _, err := reader.Get(ctx, "missing"); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("missing identity lost: %v", err)
	}
	repo.get = func(context.Context, string) (domain.SourceDocument, error) {
		t.Fatal("empty ID reached storage")
		return domain.SourceDocument{}, nil
	}
	if _, err := reader.Get(ctx, " "); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("expected invalid query, got %v", err)
	}
}

func TestRepositoryCancellationIsPreserved(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := &repositoryStub{
		get: func(ctx context.Context, _ string) (domain.SourceDocument, error) {
			return domain.SourceDocument{}, ctx.Err()
		},
		list: func(ctx context.Context, _ ports.ListOptions) ([]domain.DocumentSummary, error) {
			return nil, ctx.Err()
		},
	}
	reader := NewReader(repo)
	if _, err := reader.Get(ctx, "id"); !errors.Is(err, context.Canceled) {
		t.Fatalf("get cancellation lost: %v", err)
	}
	if _, err := reader.List(ctx, ports.ListOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("list cancellation lost: %v", err)
	}
}
