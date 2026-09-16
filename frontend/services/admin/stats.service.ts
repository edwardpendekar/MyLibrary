import { apiClient } from "@/lib/api-client";
import type { DashboardStats, User } from "@/types/api";

export const adminStatsService = {
  dashboard: () => apiClient.get<DashboardStats>("/api/v1/admin/stats"),
};

export const adminUsersService = {
  list: (cursor?: string) => apiClient.get<User[]>(`/api/v1/admin/users${cursor ? `?cursor=${cursor}` : ""}`),
  changeRole: (id: number, role: string) => apiClient.put<void>(`/api/v1/admin/users/${id}/role`, { role }),
  deactivate: (id: number) => apiClient.delete<void>(`/api/v1/admin/users/${id}`),
  activate: (id: number) => apiClient.put<void>(`/api/v1/admin/users/${id}/activate`),
};
