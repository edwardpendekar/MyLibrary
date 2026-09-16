import { apiClient } from "@/lib/api-client";

export const highlightsService = {
  listByBook: (bookId: number) => apiClient.get<number[]>(`/api/v1/books/${bookId}/highlights`),
  add: (verseId: number) => apiClient.post<void>(`/api/v1/verses/${verseId}/highlight`),
  remove: (verseId: number) => apiClient.delete<void>(`/api/v1/verses/${verseId}/highlight`),
};
