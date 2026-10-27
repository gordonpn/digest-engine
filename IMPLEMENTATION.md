# Implementation Guide

This document provides step-by-step instructions for a coding agent (or developer) to implement Digest Engine.

> **Status:** The examples below are an early implementation sketch. Follow [ADR 0001](docs/adr/0001-northflank-postgresql-and-self-hosted-sqlite.md) and [Product Direction](docs/PRODUCT_DIRECTION.md) for current deployment and product constraints. New application logic must remain independent of database drivers; the first hosted profile uses PostgreSQL, and SQLite is the future self-hosted target.

## Phase 1: Core Infrastructure (Week 1)

### 1.1 Project Setup

```bash
mkdir Digest Engine && cd Digest Engine
go mod init github.com/yourusername/Digest Engine
```

**Dependencies**:
```bash
go get github.com/mmcdole/gofeed          # RSS parsing
go get github.com/mattn/go-sqlite3        # SQLite driver
go get github.com/sqlc-dev/sqlc/cmd/sqlc  # SQL code generation
go get github.com/spf13/viper             # Config management
go get github.com/robfig/cron/v3          # Cron scheduler
go get github.com/prometheus/client_golang # Metrics
```

### 1.2 Database Schema

Create `adapters/sqlite/schema.sql`:

```sql
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    feed_id TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT,
    score INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    published_at TIMESTAMP NOT NULL,
    first_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS digest_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_name TEXT NOT NULL,
    item_id TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id)
);

CREATE INDEX IF NOT EXISTS idx_items_feed_published
    ON items(feed_id, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_items_score
    ON items(score DESC);
CREATE INDEX IF NOT EXISTS idx_digest_history_schedule
    ON digest_history(schedule_name, sent_at);
```

### 1.3 SQL Queries (for sqlc)

Create `adapters/sqlite/queries.sql`:

```sql
-- name: UpsertItem :exec
INSERT INTO items (id, feed_id, title, url, description, score, published_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    score = MAX(score, excluded.score),
    comment_count = excluded.comment_count,
    last_updated_at = CURRENT_TIMESTAMP;

-- name: GetItemsByFeedAndTimeRange :many
SELECT * FROM items
WHERE feed_id = ?
  AND published_at >= ?
  AND published_at < ?
ORDER BY score DESC
LIMIT ?;

-- name: GetItemsNotInDigest :many
SELECT i.* FROM items i
WHERE i.feed_id = ?
  AND i.published_at >= ?
  AND i.published_at < ?
  AND i.id NOT IN (
      SELECT item_id FROM digest_history
      WHERE schedule_name IN (sqlc.slice('previous_schedules'))
  )
ORDER BY i.score DESC
LIMIT ?;

-- name: RecordDigest :exec
INSERT INTO digest_history (schedule_name, item_id)
VALUES (?, ?);
```

Configure sqlc in `sqlc.yaml`:

```yaml
version: "2"
sql:
  - schema: "adapters/sqlite/schema.sql"
    queries: "adapters/sqlite/queries.sql"
    engine: "sqlite"
    gen:
      go:
        package: "sqlite"
        out: "adapters/sqlite"
```

Run: `sqlc generate`

### 1.4 Domain Models

Create `domain/feed.go`:

```go
package domain

import "time"

type Feed struct {
    ID       string
    URL      string
    Schedule string // "daily", "weekly", "monthly"
    MinScore int
    Blocklist []string // regex patterns
}

type Item struct {
    ID           string
    FeedID       string
    Title        string
    URL          string
    Description  string
    Score        int
    CommentCount int
    PublishedAt  time.Time
    FirstSeenAt  time.Time
    LastUpdatedAt time.Time
}

type Digest struct {
    Schedule  string
    Items     []Item
    Generated time.Time
}
```

## Phase 2: Adapters (Week 1-2)

### 2.1 RSS Parser Adapter

Create `adapters/rss/parser.go`:

```go
package rss

import (
    "context"
    "github.com/mmcdole/gofeed"
    "github.com/yourusername/Digest Engine/domain"
    "regexp"
    "time"
)

type Parser struct {
    gofeed *gofeed.Parser
}

func NewParser() *Parser {
    return &Parser{
        gofeed: gofeed.NewParser(),
    }
}

func (p *Parser) Parse(ctx context.Context, feed domain.Feed) ([]domain.Item, error) {
    parsed, err := p.gofeed.ParseURLWithContext(feed.URL, ctx)
    if err != nil {
        return nil, err
    }

    var items []domain.Item
    for _, item := range parsed.Items {
        // Filter by blocklist
        if p.isBlocked(item.Title, feed.Blocklist) {
            continue
        }

        items = append(items, domain.Item{
            ID:          item.GUID,
            FeedID:      feed.ID,
            Title:       item.Title,
            URL:         item.Link,
            Description: truncate(item.Description, 500),
            Score:       extractScore(item), // Implement score extraction
            PublishedAt: *item.PublishedParsed,
        })
    }

    return items, nil
}

func (p *Parser) isBlocked(title string, patterns []string) bool {
    for _, pattern := range patterns {
        if matched, _ := regexp.MatchString(pattern, title); matched {
            return true
        }
    }
    return false
}
```

**TODO for coding agent**: Implement `extractScore()` to parse Reddit karma from custom RSS fields.

### 2.2 SQLite Repository

Create `adapters/sqlite/repository.go`:

```go
package sqlite

import (
    "context"
    "database/sql"
    "github.com/yourusername/Digest Engine/domain"
    _ "github.com/mattn/go-sqlite3"
)

type Repository struct {
    db      *sql.DB
    queries *Queries // Generated by sqlc
}

func NewRepository(dbPath string) (*Repository, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // Run migrations
    if err := runMigrations(db); err != nil {
        return nil, err
    }

    return &Repository{
        db:      db,
        queries: New(db),
    }, nil
}

func (r *Repository) SaveItem(ctx context.Context, item domain.Item) error {
    return r.queries.UpsertItem(ctx, UpsertItemParams{
        ID:          item.ID,
        FeedID:      item.FeedID,
        Title:       item.Title,
        URL:         item.URL,
        Description: sql.NullString{String: item.Description, Valid: true},
        Score:       int64(item.Score),
        PublishedAt: item.PublishedAt,
    })
}

func (r *Repository) GetTopItems(ctx context.Context, feedID string, start, end time.Time, limit int) ([]domain.Item, error) {
    rows, err := r.queries.GetItemsByFeedAndTimeRange(ctx, GetItemsByFeedAndTimeRangeParams{
        FeedID:      feedID,
        PublishedAt: start,
        PublishedAt_2: end,
        Limit:       int64(limit),
    })
    // Map to domain.Item...
    return items, err
}
```

### 2.3 SMTP Notifier

Create `adapters/smtp/notifier.go`:

```go
package smtp

import (
    "fmt"
    "net/smtp"
)

type Config struct {
    Host     string
    Port     int
    Username string
    Password string
    From     string
}

type Notifier struct {
    config Config
}

func NewNotifier(config Config) *Notifier {
    return &Notifier{config: config}
}

func (n *Notifier) Send(to string, subject string, htmlBody string) error {
    auth := smtp.PlainAuth("", n.config.Username, n.config.Password, n.config.Host)

    msg := fmt.Sprintf("From: %s\r\n"+
        "To: %s\r\n"+
        "Subject: %s\r\n"+
        "MIME-Version: 1.0\r\n"+
        "Content-Type: text/html; charset=UTF-8\r\n"+
        "\r\n%s", n.config.From, to, subject, htmlBody)

    addr := fmt.Sprintf("%s:%d", n.config.Host, n.config.Port)
    return smtp.SendMail(addr, auth, n.config.From, []string{to}, []byte(msg))
}
```

## Phase 3: Application Logic (Week 2)

### 3.1 Poller with Worker Pool

Create `internal/poller/poller.go`:

```go
package poller

import (
    "context"
    "sync"
    "github.com/yourusername/Digest Engine/domain"
    "github.com/yourusername/Digest Engine/ports"
)

type Poller struct {
    parser  ports.FeedParser
    repo    ports.ItemRepository
    workers int
}

func NewPoller(parser ports.FeedParser, repo ports.ItemRepository, workers int) *Poller {
    return &Poller{
        parser:  parser,
        repo:    repo,
        workers: workers,
    }
}

func (p *Poller) Poll(ctx context.Context, feeds []domain.Feed) error {
    jobs := make(chan domain.Feed, len(feeds))
    errors := make(chan error, len(feeds))

    var wg sync.WaitGroup

    // Start workers
    for i := 0; i < p.workers; i++ {
        wg.Add(1)
        go p.worker(ctx, jobs, errors, &wg)
    }

    // Send jobs
    for _, feed := range feeds {
        jobs <- feed
    }
    close(jobs)

    // Wait for completion
    go func() {
        wg.Wait()
        close(errors)
    }()

    // Collect errors
    var errs []error
    for err := range errors {
        if err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("polling errors: %v", errs)
    }
    return nil
}

func (p *Poller) worker(ctx context.Context, jobs <-chan domain.Feed, errors chan<- error, wg *sync.WaitGroup) {
    defer wg.Done()

    for feed := range jobs {
        items, err := p.parser.Parse(ctx, feed)
        if err != nil {
            errors <- fmt.Errorf("feed %s: %w", feed.ID, err)
            continue
        }

        for _, item := range items {
            if err := p.repo.SaveItem(ctx, item); err != nil {
                errors <- fmt.Errorf("save item %s: %w", item.ID, err)
            }
        }

        errors <- nil
    }
}
```

### 3.2 Digest Generator

Create `internal/digester/digester.go`:

```go
package digester

import (
    "context"
    "fmt"
    "html/template"
    "time"
    "github.com/yourusername/Digest Engine/domain"
    "github.com/yourusername/Digest Engine/ports"
)

type Digester struct {
    repo     ports.ItemRepository
    notifier ports.Notifier
    tmpl     *template.Template
}

func NewDigester(repo ports.ItemRepository, notifier ports.Notifier) (*Digester, error) {
    tmpl, err := template.ParseFiles("templates/digest.html")
    if err != nil {
        return nil, err
    }

    return &Digester{
        repo:     repo,
        notifier: notifier,
        tmpl:     tmpl,
    }, nil
}

func (d *Digester) Generate(ctx context.Context, schedule string, feeds []domain.Feed, recipient string) error {
    var allItems []domain.Item

    start, end := d.getTimeRange(schedule)

    for _, feed := range feeds {
        if feed.Schedule != schedule {
            continue
        }

        items, err := d.repo.GetTopItems(ctx, feed.ID, start, end, 10)
        if err != nil {
            return fmt.Errorf("get items for %s: %w", feed.ID, err)
        }

        allItems = append(allItems, items...)
    }

    if len(allItems) == 0 {
        return nil // No digest to send
    }

    htmlBody, err := d.renderTemplate(schedule, allItems)
    if err != nil {
        return err
    }

    subject := fmt.Sprintf("Digest Engine %s Digest - %s", schedule, time.Now().Format("Jan 2"))
    return d.notifier.Send(recipient, subject, htmlBody)
}

func (d *Digester) getTimeRange(schedule string) (start, end time.Time) {
    end = time.Now()
    switch schedule {
    case "daily":
        start = end.AddDate(0, 0, -1)
    case "weekly":
        start = end.AddDate(0, 0, -7)
    case "monthly":
        start = end.AddDate(0, -1, 0)
    }
    return
}
```

## Phase 4: CLI & Orchestration (Week 2-3)

### 4.1 Main CLI

Create `cmd/Digest Engine/main.go`:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/viper"
    "github.com/yourusername/Digest Engine/internal/poller"
    "github.com/yourusername/Digest Engine/internal/digester"
    "github.com/yourusername/Digest Engine/adapters/rss"
    "github.com/yourusername/Digest Engine/adapters/sqlite"
    "github.com/yourusername/Digest Engine/adapters/smtp"
)

func main() {
    // Load config
    viper.SetConfigName("config")
    viper.AddConfigPath(".")
    if err := viper.ReadInConfig(); err != nil {
        log.Fatal(err)
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        log.Fatal(err)
    }

    // Initialize adapters
    repo, err := sqlite.NewRepository("data/Digest Engine.db")
    if err != nil {
        log.Fatal(err)
    }

    parser := rss.NewParser()
    notifier := smtp.NewNotifier(config.SMTP)

    // Initialize services
    p := poller.NewPoller(parser, repo, 10)
    d, err := digester.NewDigester(repo, notifier)
    if err != nil {
        log.Fatal(err)
    }

    // Run based on command
    if len(os.Args) < 2 {
        log.Fatal("Usage: Digest Engine [poll|digest|serve]")
    }

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    switch os.Args {
    case "poll":
        if err := p.Poll(ctx, config.Feeds); err != nil {
            log.Fatal(err)
        }
    case "digest":
        schedule := os.Args // "daily", "weekly", "monthly"
        if err := d.Generate(ctx, schedule, config.Feeds, config.SMTP.To); err != nil {
            log.Fatal(err)
        }
    case "serve":
        // Implement scheduler (see Phase 5)
    }
}
```

## Phase 5: Production Features (Week 3+)

### 5.1 Prometheus Metrics

Add to each component:

```go
var (
    feedPollsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "Digest Engine_feed_polls_total",
            Help: "Total number of feed polls",
        },
        []string{"feed_id", "status"},
    )
)

func init() {
    prometheus.MustRegister(feedPollsTotal)
}

// In poller
feedPollsTotal.WithLabelValues(feed.ID, "success").Inc()
```

### 5.2 Kubernetes Manifests

See `DEPLOYMENT.md` for full K8s YAML.

## Testing Strategy

```go
// Unit tests: domain logic
func TestFeed_Validate(t *testing.T) { ... }

// Integration tests: adapters
func TestSQLiteRepository_SaveItem(t *testing.T) {
    repo, _ := sqlite.NewRepository(":memory:")
    // ...
}

// E2E tests: full flow
func TestPollAndDigest(t *testing.T) {
    // Use test fixtures
}
```

## Next Steps

1. Implement score extraction for Reddit/HN (parse custom fields)
2. Add HTML email template (`templates/digest.html`)
3. Implement scheduler with `robfig/cron`
4. Add observability (metrics endpoint, structured logging)
5. Create Kubernetes manifests
