import { apiClient } from "@/lib/api-client";
import type { Bookmark } from "@/types/api";

export interface CreateBookmarkInput {
  book_id: number;
  chapter_id?: number;
  verse_id?: number;
  pdf_page?: number;
  label?: string;
}

export interface SaveLastPositionInput {
  book_id: number;
  chapter_id?: number;
  verse_id?: number;
  pdf_page?: number;
}

export const bookmarksService = {
  listByBook: (bookId: number) => apiClient.get<Bookmark[]>(`/api/v1/books/${bookId}/bookmarks`),
  create: (input: CreateBookmarkInput) => apiClient.post<Bookmark>("/api/v1/bookmarks", input),
  remove: (id: number) => apiClient.delete<void>(`/api/v1/bookmarks/${id}`),
  saveLastPosition: (input: SaveLastPositionInput) =>
    apiClient.put<void>("/api/v1/bookmarks/last-position", input),
  lastPosition: (bookId: number) => apiClient.get<Bookmark | null>(`/api/v1/books/${bookId}/last-position`),
};
