import { apiClient } from "@/lib/api-client";
import type { Book, SearchVerseHit } from "@/types/api";

export const searchService = {
  verses: (q: string, opts: { bookId?: number; cursor?: string; limit?: number } = {}) => {
    const params = new URLSearchParams({ q });
    if (opts.bookId) params.set("book_id", String(opts.bookId));
    if (opts.cursor) params.set("cursor", opts.cursor);
    if (opts.limit) params.set("limit", String(opts.limit));
    return apiClient.get<SearchVerseHit[]>(`/api/v1/search/verses?${params.toString()}`);
  },
  books: (q: string, opts: { cursor?: string; limit?: number } = {}) => {
    const params = new URLSearchParams({ q });
    if (opts.cursor) params.set("cursor", opts.cursor);
    if (opts.limit) params.set("limit", String(opts.limit));
    return apiClient.get<Book[]>(`/api/v1/search/books?${params.toString()}`);
  },
};
