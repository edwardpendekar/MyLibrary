# Reader flow

Two independent readers share a book: the **verse reader** (chapter/section/verse
text, EN/ID side-by-side) and the **PDF reader** (`pdf.js`-based). Both persist
"last read position" the same way.

```mermaid
flowchart TD
    A[/books/slug/] -->|Read button| B[/books/slug/read/1/]
    A -->|Open PDF button| C[/books/slug/pdf/]

    subgraph Verse Reader
    B --> D[ChapterList: GET /books/:id/chapters]
    B --> E[VerseReader: GET /books/:id/chapters/:number]
    B --> F[RightPanel: bookmarks / notes / in-book search]
    E -->|on mount, if logged in| G[PUT /bookmarks/last-position chapter_id]
    end

    subgraph PDF Reader
    C --> H[pdfjs-dist loads book.pdf_url]
    H --> I[canvas render current page]
    C -->|on page change, if logged in| G
    end

    J[Return visit to /books/slug/read or /pdf] -->|GET /books/:id/last-position| K[resume at saved chapter/page]
```

## Verse reader

- **Data shape.** `GET /books/:id/chapters/:number` returns the chapter, its
  `sections` (heading groups), and all `verses` in one call — the frontend
  groups verses under their section client-side (`groupBySection` in
  `verse-reader.tsx`) rather than issuing N+1 requests.
- **Per-verse actions** (copy, highlight, bookmark, note, share) live in
  `VerseItem`. Highlights are device-local (`store/highlights-store.ts`,
  persisted to `localStorage`) since there is no `highlights` table in the
  schema yet — bookmarks and notes are the account-synced equivalents.
- **Font size / EN·ID·both toggle** are user preferences persisted to
  `localStorage` (`store/reader-preferences-store.ts`), not synced to the
  account — they are reading-device UI state, not content.
- **In-book search** (`RightPanel`'s Search tab) calls the same
  `/search/verses` endpoint as the global search page, scoped with
  `book_id`.

## PDF reader

- Built directly on `pdfjs-dist` (no wrapper library) — see
  `features/reader/components/pdf-viewer.tsx` and `hooks/use-pdf-document.ts`.
  The worker script is copied into `public/pdfjs/` at install time
  (`scripts/copy-pdf-worker.js`) so it's served same-origin, versioned to
  match the installed `pdfjs-dist` exactly.
- **Zoom** re-renders the current page's canvas at a new `scale`. **Search**
  walks every page's extracted text client-side and lets the reader jump to
  matching pages (no in-page highlight overlay — a deliberate scope cut, see
  the comment in `pdf-viewer.tsx`). **Dark mode** is a CSS `invert()` filter
  on the canvas, not a PDF re-render. **Fullscreen** uses the standard
  `Fullscreen` browser API on the viewer's container.

## "Remember last page"

Implemented as a single `is_auto=true` row per `(user_id, book_id)` in
`bookmarks`, enforced by a partial unique index
(`uq_bookmarks_auto_position`) and upserted via `PUT
/bookmarks/last-position`. The verse reader saves at chapter granularity (on
chapter mount); the PDF reader saves on every page change. Both read it back
via `GET /books/:id/last-position` to resume on the next visit.
