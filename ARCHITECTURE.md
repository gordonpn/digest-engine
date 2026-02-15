# Architecture

## Design Philosophy

Digest Engine is built with three principles:

1. **Fundamentals Over Magic**: Uses Go stdlib and minimal abstractions to ensure understanding of underlying mechanisms
2. **Production-Ready Patterns**: Implements enterprise architecture patterns (Hexagonal, Event-Driven)
3. **Observable by Default**: All components expose metrics and structured logs

## High-Level Design

### The "Store-Then-Digest" Pattern

Unlike traditional RSS readers that fetch-on-demand, Digest Engine uses a two-phase approach:

```
Phase 1 (Continuous): Poller → Store
  - Runs hourly
  - Captures items when scores are low
  - Updates scores as items gain traction
  - Result: High-fidelity dataset of "what was popular"

Phase 2 (Scheduled): Store → Digest → Email
  - Runs on user schedule (daily/weekly/monthly)
  - Queries stored items by timeframe
  - Ranks by final score (not initial score)
  - Filters out items already sent in previous digests
```

**Why This Matters**: A post on Reddit might have 2 upvotes at hour 1, then 500 upvotes by hour 12. Traditional readers miss this. Digest Engine captures the trajectory.

## Component Architecture (Hexagonal/Ports & Adapters)

```
Digest Engine/
├── domain/              # Business logic (no external dependencies)
│   ├── feed.go          # Feed entity, validation
│   ├── item.go          # Item entity
│   └── digest.go        # Digest generation logic
│
├── ports/               # Interfaces (contracts)
│   ├── repository.go    # type ItemRepository interface
│   ├── notifier.go      # type Notifier interface
│   └── parser.go        # type FeedParser interface
│
├── adapters/            # External integrations
│   ├── sqlite/          # ItemRepository implementation
│   │   ├── migrations/
│   │   └── queries.sql  # Used by sqlc
│   ├── smtp/            # Notifier implementation
│   └── rss/             # FeedParser implementation (wraps gofeed)
│
├── internal/            # Application logic
│   ├── poller/          # Polling orchestration
│   │   └── worker_pool.go
│   ├── digester/        # Digest generation
│   │   └── template.go
│   └── scheduler/       # Cron management
│
└── cmd/
    └── Digest Engine/
        └── main.go      # CLI entry point
```

## Data Model

### Core Tables

```sql
-- items: Stores all feed items
CREATE TABLE items (
    id TEXT PRIMARY KEY,              -- URL or GUID
    feed_id TEXT NOT NULL,            -- Links to config
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT,                 -- First 500 chars or og:description
    score INTEGER DEFAULT 0,          -- Upvotes/points
    comment_count INTEGER DEFAULT 0,  -- For future use
    published_at TIMESTAMP NOT NULL,
    first_seen_at TIMESTAMP NOT NULL, -- When we first polled it
    last_updated_at TIMESTAMP NOT NULL,
    UNIQUE(feed_id, url)
);

-- digest_history: Tracks what was sent
CREATE TABLE digest_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_name TEXT NOT NULL,      -- e.g., "weekly"
    item_id TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL,
    FOREIGN KEY (item_id) REFERENCES items(id)
);

CREATE INDEX idx_items_feed_published ON items(feed_id, published_at);
CREATE INDEX idx_items_score ON items(score DESC);
```

## Concurrency Design

### Poller: Worker Pool Pattern

```go
// internal/poller/pool.go
type Pool struct {
    workers   int
    jobs      chan Feed
    results   chan Result
    parser    ports.FeedParser
    repo      ports.ItemRepository
}

func (p *Pool) Poll(feeds []domain.Feed) error {
    // 1. Spawn worker goroutines (limit concurrency)
    for i := 0; i < p.workers; i++ {
        go p.worker()
    }

    // 2. Send jobs to workers
    for _, feed := range feeds {
        p.jobs <- feed
    }
    close(p.jobs)

    // 3. Collect results
    var errs []error
    for range feeds {
        if r := <-p.results; r.Error != nil {
            errs = append(errs, r.Error)
        }
    }

    return errors.Join(errs...)
}
```

**Why**: Limits concurrent HTTP requests (prevents rate-limiting). Teaches channel semantics and graceful error aggregation.

### Scheduler: Ticker + Select Pattern

```go
// internal/scheduler/scheduler.go
func (s *Scheduler) Run(ctx context.Context) error {
    pollTicker := time.NewTicker(1 * time.Hour)
    defer pollTicker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-pollTicker.C:
            s.poller.Poll(s.feeds)
        case t := <-s.cronChan:
            s.digester.Generate(t.Schedule)
        }
    }
}
```

**Why**: Single event loop handles multiple timers. Teaches context cancellation and graceful shutdown.

## Observability

### Metrics (Prometheus)

```go
var (
    feedFetchTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "Digest Engine_fetch_total"},
        []string{"feed_id", "status"},
    )

    itemsIngestedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "Digest Engine_items_ingested_total"},
        []string{"feed_id"},
    )

    digestSentTimestamp = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{Name: "Digest Engine_digest_sent_timestamp"},
        []string{"schedule"},
    )
)
```

**Alerting Rules**:
- `Digest Engine_fetch_total{status="error"} > 5` (Feed down)
- `time() - Digest Engine_digest_sent_timestamp > 86400` (Digest didn't send)

### Logging (Structured JSON)

```go
log.Info("feed polled",
    "feed_id", feed.ID,
    "items_new", newCount,
    "items_updated", updatedCount,
    "duration_ms", elapsed.Milliseconds(),
)
```

## Configuration Management

Uses YAML with Viper for:
- Hot-reloading (watches config file for changes)
- Environment variable overrides (`Digest Engine_SMTP_PASSWORD`)
- Validation on startup

```go
type Config struct {
    Feeds     []FeedConfig     `mapstructure:"feeds"`
    Schedules map[string]string `mapstructure:"schedules"`
    SMTP      SMTPConfig       `mapstructure:"smtp"`
}

func (c *Config) Validate() error {
    // Ensure all feed schedules exist in schedules map
    // Validate cron syntax
    // Check SMTP connectivity
}
```

## Deployment Patterns

### Development
- Single binary (`go run cmd/Digest Engine/main.go serve`)
- SQLite file in `./data/`

### Production (Kubernetes)
- **Poller**: Deployment with 1 replica (leader election not needed, SQLite is single-writer)
- **Digester**: CronJob per schedule (daily, weekly, monthly)
- **Storage**: PersistentVolumeClaim for SQLite file
- **Secrets**: SMTP credentials in K8s Secret

## Trade-offs & Decisions

| Decision | Why | Trade-off |
|----------|-----|-----------|
| SQLite vs Postgres | Simpler ops, single binary | No multi-writer (fine for this use case) |
| Stdlib HTTP vs Gin | Learn middleware patterns | More boilerplate |
| Worker pool vs Library | Understand concurrency | Reinventing wheel |
| Embedded templates | No external dependencies | Harder to hot-reload |
| SMTP vs API | Universal support | Requires SMTP credentials |

## Future Enhancements

- [ ] Local LLM summarization (Ollama integration)
- [ ] Web UI for browsing archive
- [ ] Webhook receiver for "Save to Notion"
- [ ] Multi-user support (user table, per-user configs)
