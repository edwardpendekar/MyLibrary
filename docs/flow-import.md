# Import flow

End-to-end path for turning an Excel/CSV file into Book → Chapter → Section →
Verse rows, as implemented in `backend/internal/service/import_service.go`
and driven by the admin UI at `frontend/app/[locale]/admin/import/page.tsx`.

```mermaid
sequenceDiagram
    actor Admin
    participant UI as Admin UI
    participant API as Go API
    participant DB as PostgreSQL

    Admin->>UI: choose file (.xlsx/.xls/.csv)
    UI->>API: POST /admin/import/upload (multipart)
    API->>API: store file (storage.Provider)
    API->>API: create import_logs row (status=validating)
    API->>API: stream-parse file: validate header,<br/>count rows, collect sample + errors
    API->>DB: update import_logs (status=ready, total_rows)
    API-->>UI: preview (counts, sample rows, validation errors)

    Admin->>UI: pick mode (insert/upsert), click "Start import"
    UI->>API: POST /admin/import/:id/commit {mode}
    API-->>UI: 200 (job accepted, runs in background)

    loop every ~2000-row batch
        API->>API: find/create Book, Chapter, Section (cached in-memory)
        API->>DB: COPY batch into temp staging table
        API->>DB: INSERT ... SELECT ... ON CONFLICT (merge)
        API->>DB: update import_logs (processed_rows, counts)
    end

    UI->>API: GET /admin/import/:id/stream (SSE)
    API-->>UI: progress event every 1s until terminal status

    alt success
        API->>DB: recalculate chapters_count/verses_count per touched book
        API->>API: invalidate book + search caches
        API->>DB: import_logs.status = completed
    else failure
        API->>DB: soft-delete books newly created by this run
        API->>DB: import_logs.status = failed, error_message set
    end
```

## Design decisions

- **Streaming parse.** `excelize`'s `Rows()` iterator (xlsx) and a buffered
  `encoding/csv.Reader` (csv) never load the whole file into memory — required
  to make a 1M+ row import feasible at all.
- **Batched COPY, not one giant transaction.** Each batch is its own
  transaction (COPY into a temp table + one set-based merge). This is what
  makes 1M+ row throughput possible, but it means failure recovery is a
  **compensating action**, not a database ROLLBACK — see the trade-off
  documented at the top of `import_service.go`.
- **Duplicate detection.** `insert` mode uses `ON CONFLICT (chapter_id,
  number) DO NOTHING` (existing verses are skipped and counted);` upsert` mode
  uses `DO UPDATE` (existing verses are overwritten and counted separately).
- **Sections.** The Excel `title_en`/`title_id` columns are a heading that
  repeats across many consecutive verse rows (e.g. "Creation" for Genesis
  1:1-1:5). Rather than duplicate that text on every verse, the importer
  normalizes it into a `sections` row the first time it sees that heading in
  a chapter, and extends `end_verse_number` as more matching rows arrive.
- **Rollback scope.** On failure, only books this run itself created are
  removed (cascading to their chapters/sections/verses via FK `ON DELETE
  CASCADE`). If the run was upserting into an already-existing book, verses
  merged before the failure are left in place — documented as a known
  trade-off, not silently swept under the rug.
