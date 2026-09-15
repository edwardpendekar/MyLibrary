import { apiClient } from "@/lib/api-client";
import type { Book } from "@/types/api";

export interface UpsertBookInput {
  title: string;
  author?: string;
  description?: string;
  language_id?: number;
  category_id?: number;
  year?: number;
  isbn?: string;
  slug?: string;
  status?: "draft" | "published" | "archived";
}

export const adminBooksService = {
  list: (params: { cursor?: string; q?: string } = {}) => {
    const search = new URLSearchParams();
    if (params.cursor) search.set("cursor", params.cursor);
    if (params.q) search.set("q", params.q);
    const qs = search.toString();
    return apiClient.get<Book[]>(`/api/v1/admin/books${qs ? `?${qs}` : ""}`);
  },
  get: (id: number) => apiClient.get<Book>(`/api/v1/admin/books/${id}`),
  create: (input: UpsertBookInput) => apiClient.post<Book>("/api/v1/admin/books", input),
  update: (id: number, input: UpsertBookInput) => apiClient.put<Book>(`/api/v1/admin/books/${id}`, input),
  remove: (id: number) => apiClient.delete<void>(`/api/v1/admin/books/${id}`),
  uploadCover: (id: number, file: File) => {
    const form = new FormData();
    form.set("file", file);
    return apiClient.upload<{ cover_url: string }>(`/api/v1/admin/books/${id}/cover`, form);
  },
  uploadPdf: (id: number, file: File) => {
    const form = new FormData();
    form.set("file", file);
    return apiClient.upload<{ pdf_url: string }>(`/api/v1/admin/books/${id}/pdf`, form);
  },
};
