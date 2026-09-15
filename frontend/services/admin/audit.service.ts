import { apiClient } from "@/lib/api-client";
import type { AuditLog } from "@/types/api";

export const adminAuditService = {
  list: (cursor?: string) => apiClient.get<AuditLog[]>(`/api/v1/admin/audit-logs${cursor ? `?cursor=${cursor}` : ""}`),
};
