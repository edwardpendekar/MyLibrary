// Mirrors backend/internal/dto — kept in sync by hand. If this drifts from the
// Go structs, prefer trusting the Go source (backend/internal/dto/*.go).

export interface ApiEnvelope<T> {
  success: boolean;
  data: T;
  meta?: PageMeta;
  error?: { code: string; message: string; fields?: Record<string, string> };
}

export interface PageMeta {
  next_cursor?: string;
  has_more: boolean;
  limit: number;
}

export interface User {
  id: number;
  name: string;
  email: string;
  role: "admin" | "editor" | "user" | "guest";
  is_active: boolean;
  created_at: string;
}

export interface Language {
  id: number;
  code: string;
  name: string;
  native_name: string;
}

export interface Category {
  id: number;
  slug: string;
  name_en: string;
  name_id: string;
}

export interface Book {
  id: number;
  slug: string;
  title: string;
  author: string;
  description: string;
  language?: Language | null;
  category?: Category | null;
  year?: number | null;
  isbn?: string | null;
  cover_url?: string | null;
  status: "draft" | "published" | "archived";
  chapters_count: number;
  verses_count: number;
  view_count: number;
  has_pdf: boolean;
  pdf_url?: string | null;
  is_favorite?: boolean;
  created_at: string;
}

export interface Chapter {
  id: number;
  number: number;
  title_en?: string | null;
  title_id?: string | null;
  verses_count: number;
}

export interface Section {
  id: number;
  title_en?: string | null;
  title_id?: string | null;
  start_verse_number: number;
  end_verse_number: number;
}

export interface Verse {
  id: number;
  number: number;
  text_en?: string | null;
  text_id?: string | null;
  section_id?: number | null;
}

export interface ChapterContent {
  chapter: Chapter;
  sections: Section[];
  verses: Verse[];
}

export interface Bookmark {
  id: number;
  book_id: number;
  chapter_id?: number | null;
  verse_id?: number | null;
  pdf_page?: number | null;
  label?: string | null;
  created_at: string;
}

export interface Note {
  id: number;
  book_id: number;
  verse_id?: number | null;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface SearchVerseHit {
  book_id: number;
  book_title: string;
  book_slug: string;
  chapter_id?: number | null;
  chapter_number?: number | null;
  verse_id?: number | null;
  verse_number?: number | null;
  snippet: string;
  rank: number;
}

export interface ImportLog {
  id: number;
  filename: string;
  status:
    | "pending"
    | "validating"
    | "ready"
    | "importing"
    | "completed"
    | "failed"
    | "rolled_back";
  mode: "insert" | "upsert";
  total_rows: number;
  processed_rows: number;
  books_created: number;
  chapters_created: number;
  sections_created: number;
  verses_inserted: number;
  verses_updated: number;
  verses_skipped: number;
  error_message?: string | null;
  started_at?: string | null;
  finished_at?: string | null;
  created_at: string;
}

export interface ImportPreview {
  import_log_id: number;
  status: string;
  total_rows: number;
  distinct_books: number;
  distinct_chapters: number;
  sample_rows: Array<{
    row_number: number;
    book: string;
    chapter: number;
    verse: number;
    text_en: string;
    text_id: string;
    title_en: string;
    title_id: string;
  }>;
  validation_errors: Array<{ row_number: number; message: string }>;
}

export interface DashboardStats {
  total_books: number;
  published_books: number;
  total_chapters: number;
  total_verses: number;
  total_users: number;
  imports_last_30_day: number;
}

export interface AuditLog {
  id: number;
  user_id?: number | null;
  action: string;
  entity_type: string;
  entity_id?: string | null;
  ip_address: string;
  user_agent: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}
