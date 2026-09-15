import { apiClient } from "@/lib/api-client";
import type { Note } from "@/types/api";

export const notesService = {
  listByBook: (bookId: number) => apiClient.get<Note[]>(`/api/v1/books/${bookId}/notes`),
  create: (input: { book_id: number; verse_id?: number; content: string }) =>
    apiClient.post<Note>("/api/v1/notes", input),
  update: (id: number, content: string) => apiClient.put<Note>(`/api/v1/notes/${id}`, { content }),
  remove: (id: number) => apiClient.delete<void>(`/api/v1/notes/${id}`),
};
