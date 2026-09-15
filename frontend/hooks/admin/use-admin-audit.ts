"use client";

import { useQuery } from "@tanstack/react-query";
import { adminAuditService } from "@/services/admin/audit.service";

export function useAuditLogs() {
  return useQuery({
    queryKey: ["admin", "audit-logs"],
    queryFn: () => adminAuditService.list(),
    select: (res) => res.data,
  });
}
