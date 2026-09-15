import { apiClient } from "@/lib/api-client";
import type { Category, Language } from "@/types/api";

export const adminLanguagesService = {
  list: () => apiClient.get<Language[]>("/api/v1/admin/languages"),
  create: (input: { code: string; name: string; native_name: string }) =>
    apiClient.post<Language>("/api/v1/admin/languages", input),
  update: (id: number, input: { code: string; name: string; native_name: string; is_active: boolean }) =>
    apiClient.put<Language>(`/api/v1/admin/languages/${id}`, input),
  remove: (id: number) => apiClient.delete<void>(`/api/v1/admin/languages/${id}`),
};

export const adminCategoriesService = {
  list: () => apiClient.get<Category[]>("/api/v1/admin/categories"),
  create: (input: { slug?: string; name_en: string; name_id: string; description?: string }) =>
    apiClient.post<Category>("/api/v1/admin/categories", input),
  update: (id: number, input: { slug: string; name_en: string; name_id: string; description?: string }) =>
    apiClient.put<Category>(`/api/v1/admin/categories/${id}`, input),
  remove: (id: number) => apiClient.delete<void>(`/api/v1/admin/categories/${id}`),
};
