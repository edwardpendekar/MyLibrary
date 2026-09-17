import { apiClient } from "@/lib/api-client";
import type { ImportLog, ImportPreview } from "@/types/api";

export const adminImportService = {
  upload: (file: File) => {
    const form = new FormData();
    form.set("file", file);
    return apiClient.upload<ImportPreview>("/api/v1/admin/import/upload", form);
  },
  translate: (input: { bookId: number; chapterNumber: number; titleEn: string; bodyEn: string }) =>
    apiClient.post<ImportLog>("/api/v1/admin/import/translate", {
      book_id: input.bookId,
      chapter_number: input.chapterNumber,
      title_en: input.titleEn,
      body_en: input.bodyEn,
    }),
  preview: (id: number) => apiClient.get<ImportPreview>(`/api/v1/admin/import/${id}/preview`),
  commit: (id: number, mode: "insert" | "upsert") =>
    apiClient.post<ImportLog>(`/api/v1/admin/import/${id}/commit`, { mode }),
  status: (id: number) => apiClient.get<ImportLog>(`/api/v1/admin/import/${id}`),
  history: (cursor?: string) =>
    apiClient.get<ImportLog[]>(`/api/v1/admin/import${cursor ? `?cursor=${cursor}` : ""}`),
  streamUrl: (id: number) => {
    const origin =
      process.env.NEXT_PUBLIC_BACKEND_ORIGIN ?? "http://localhost:8080";
    return `${origin}/api/v1/admin/import/${id}/stream`;
  },
};
