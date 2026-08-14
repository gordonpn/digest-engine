# Zero-Fluff & Anti-Clickbait Enrichment Design

This document details the architecture, design patterns, and implementation strategy for the **Zero-Fluff & Anti-Clickbait ("De-Baiter") Enrichment Pipeline** in Digest Engine.

---

## 1. Problem Statement & Objectives

### The Problem
* **Clickbait Headlines**: Titles use curiosity gaps, emotional triggers, and omitted subjects (*"Apple just quietly killed this feature"*).
* **SEO Padding & Buried Ledes**: Articles often stretch a 2-sentence fact into 1,000+ words of historical backstory, rhetorical questions, and repetitive filler.
* **Information Overload**: Readers spend 5–10 minutes reading an article only to discover the actual announcement takes 10 seconds to understand.

### The Objectives
1. **Spoiler / Factual Titles**: Rewrite vague or clickbait headlines into direct, factual titles containing the subject and outcome.
2. **BLUF (Bottom Line Up Front)**: Extract a 1–2 sentence summary stating the core news, decision, or discovery.
3. **Key Takeaways & Evidence**: Surface 2–3 concise bullet points with hard facts, numbers, benchmarks, or direct quotes.
4. **Fluff Reduction Metrics**: Calculate estimated reading time saved and filler percentage.
5. **Zero Waste (JIT Processing)**: Enrich only top-ranked items selected for digests, not every polled item.

---

## 2. Digest Presentation Format

Each enriched item rendered in the digest email displays as a high-signal executive card:

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ [Source: r/golang] ⚡ ZERO-FLUFF DIGEST                                      │
│                                                                              │
│ 🎯 Factual Title: Go 1.24 Stabilizes Weak Pointers & Generic Type Aliases    │
│    (Original: "You won't believe what just landed in the latest Go release") │
│                                                                              │
│ 📌 BLUF (Bottom Line Up Front):                                              │
│    Go 1.24 introduces stdlib weak pointers (`weak.Pointer`) to prevent memory│
│    leaks in long-lived caches, and enables full generic type aliases.        │
│                                                                              │
│ 🔍 Key Takeaways:                                                            │
│    • `weak.Pointer` eliminates custom finalizer hacks for caching structures.│
│    • Generic type aliases (`type Set[T] = other.Set[T]`) now fully supported.│
│    • Benchmark shows ~3% reduction in GC overhead for cache-heavy workloads. │
│                                                                              │
│ ⏱️ Saved ~4 min read (78% fluff removed) • Score: 412 pts • [Read Source]    │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. System Architecture & Lifecycle

### Just-in-Time (JIT) Enrichment Pipeline
Enrichment is executed **only when a digest schedule triggers** (e.g., daily, weekly, monthly) for the top $N$ items that qualify for delivery.

```
┌──────────────┐
│ Poller       │ ──> [items table] (raw RSS title, URL, score, timestamps)
└──────────────┘
                           │
                           ▼ (Scheduled Cron: Daily / Weekly / Monthly)
┌──────────────┐
│ Digester     │ ──> Query top N candidate items by score
└──────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│ Enrichment Service (Worker Pool)                            │
│                                                             │
│  1. Check SQLite Cache (`is_enriched == true`) ─────────┐   │
│     (If cached, skip external fetching/LLM)             │   │
│                                                         │   │
│  2. ports.ContentExtractor                              │   │
│     (Fetch article HTML -> Extract clean text body)     │   │
│                                                         │   │
│  3. ports.Enricher (LLM Client)                         │   │
│     (De-bait title + extract BLUF + bullet takeaways)   │   │
│                                                         │   │
│  4. ports.ItemRepository                                │   │
│     (Persist enriched metadata to SQLite cache) <───────┘   │
└─────────────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────┐
│ Template     │ ──> Render zero-fluff HTML cards ──> Send Email (SMTP)
└──────────────┘
```

---

## 4. Hexagonal Architecture Contracts

### 4.1 Domain Entities (`domain/`)

```go
// domain/enrichment.go
package domain

type EnrichmentResult struct {
    CleanTitle    string   `json:"clean_title"`
    BLUF          string   `json:"bluf"`
    KeyTakeaways  []string `json:"key_takeaways"`
    FluffScore    int      `json:"fluff_score"`    // 0 to 100%
    EstimatedMins int      `json:"estimated_mins"` // Read time saved in minutes
}

// Extended domain.Item in domain/item.go
type Item struct {
    ID            string
    FeedID        string
    Title         string
    URL           string
    Description   string
    Score         int
    CommentCount  int
    PublishedAt   time.Time
    FirstSeenAt   time.Time
    LastUpdatedAt time.Time

    // Enriched fields
    CleanTitle    string
    BLUF          string
    KeyTakeaways  []string
    FluffScore    int
    IsEnriched    bool
}
```

### 4.2 Ports (`ports/`)

```go
// ports/extractor.go
package ports

import "context"

type ContentExtractor interface {
    // Extract scrapes the target URL and returns sanitized article plain text/markdown.
    Extract(ctx context.Context, url string) (string, error)
}

// ports/enricher.go
package ports

import (
    "context"
    "github.com/yourusername/digest-engine/domain"
)

type Enricher interface {
    // Enrich processes the raw headline and extracted content into structured data.
    Enrich(ctx context.Context, title, content string) (*domain.EnrichmentResult, error)
}
```

### 4.3 Adapters (`adapters/`)
* **`adapters/extractor/readability`**: Scrapes URL, strips navigation/ads using Readability algorithm with timeout handling.
* **`adapters/enricher/ollama`**: Connects to local Ollama instance (e.g. `llama3.2:3b`, `qwen2.5:3b`).
* **`adapters/enricher/gemini`**: Connects to Gemini 2.5 Flash API via standard HTTP / SDK.

---

## 5. LLM Prompt & Output Schema Contract

```json
{
  "system": "You are an objective, no-fluff intelligence analyst. Given an article title and content, your job is to eliminate clickbait, resolve vague pronouns, cut SEO filler, and extract the bottom-line facts.",
  "rules": [
    "Clean title must state the subject, action, and key detail. If the original title asked a question, answer it.",
    "If the original title is already concise and factual, keep it.",
    "BLUF must be 1-2 sentences maximum, strictly summarizing the core news or decision with zero preamble.",
    "Key takeaways must contain 2-3 concrete facts, numbers, benchmarks, or direct quotes.",
    "Estimate fluff_percentage (0-100) based on how much filler and SEO preamble the article had."
  ],
  "response_format": {
    "type": "json_object",
    "schema": {
      "clean_title": "string",
      "bluf": "string",
      "key_takeaways": ["string"],
      "fluff_percentage": 0,
      "estimated_mins_saved": 0
    }
  }
}
```

---

## 6. Database Schema Evolution

SQLite schema additions to persist enrichment without re-querying LLMs:

```sql
-- Migration: Add enrichment columns to items
ALTER TABLE items ADD COLUMN clean_title TEXT;
ALTER TABLE items ADD COLUMN bluf TEXT;
ALTER TABLE items ADD COLUMN key_takeaways TEXT; -- JSON serialized array
ALTER TABLE items ADD COLUMN fluff_score INTEGER DEFAULT 0;
ALTER TABLE items ADD COLUMN is_enriched BOOLEAN DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_items_enriched ON items(is_enriched);
```

---

## 7. Resilience & Fallback Matrix

| Scenario | Fallback Behavior |
| :--- | :--- |
| **Paywall / Scrape Block (403, Cloudflare)** | Fall back to RSS `description` or summary snippet as the content input for the enricher. |
| **LLM Offline / Timeout / Rate Limit** | Mark item `is_enriched = false` and render raw RSS title and description without breaking the digest. |
| **Non-Article URLs (GitHub repos, PDFs, Videos)** | Detect URL pattern; use domain-specific extractors (e.g., GitHub README via API) or fallback to RSS description. |
| **Already Factual Headline** | The LLM preserves the original headline if no curiosity gap or clickbait is detected. |
