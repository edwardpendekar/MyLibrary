import { apiClient } from "@/lib/api-client";
import type { Book, Chapter, ChapterContent } from "@/types/api";

export interface ListBooksParams {
  cursor?: string;
  limit?: number;
  category_id?: number;
  language_id?: number;
  q?: string;
}

function toQuery(params: object): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") search.set(key, String(value));
  }
  const qs = search.toString();
  return qs ? `?${qs}` : "";
}

export const booksService = {
  list: (params: ListBooksParams = {}) =>
    apiClient.get<Book[]>(`/api/v1/books${toQuery(params)}`),
  popular: (limit = 10) => apiClient.get<Book[]>(`/api/v1/books/popular?limit=${limit}`),
  bySlug: (slug: string) => apiClient.get<Book>(`/api/v1/books/by-slug/${slug}`),
  chapters: (bookId: number) => apiClient.get<Chapter[]>(`/api/v1/books/${bookId}/chapters`),
  chapterContent: (bookId: number, chapterNumber: number) =>
    apiClient.get<ChapterContent>(`/api/v1/books/${bookId}/chapters/${chapterNumber}`),
};
