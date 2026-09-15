import { apiClient } from "@/lib/api-client";
import type { Book } from "@/types/api";

export const favoritesService = {
  add: (bookId: number) => apiClient.post<void>(`/api/v1/books/${bookId}/favorite`),
  remove: (bookId: number) => apiClient.delete<void>(`/api/v1/books/${bookId}/favorite`),
  list: (cursor?: string) =>
    apiClient.get<Book[]>(`/api/v1/me/favorites${cursor ? `?cursor=${cursor}` : ""}`),
};
