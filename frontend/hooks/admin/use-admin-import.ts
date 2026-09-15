"use client";

import { useEffect, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { adminImportService } from "@/services/admin/import.service";
import type { ImportLog } from "@/types/api";

export function useImportHistory() {
  return useQuery({
    queryKey: ["admin", "import", "history"],
    queryFn: () => adminImportService.history(),
    select: (res) => res.data,
  });
}

export function useUploadImportFile() {
  return useMutation({
    mutationFn: (file: File) => adminImportService.upload(file),
  });
}

export function useCommitImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, mode }: { id: number; mode: "insert" | "upsert" }) => adminImportService.commit(id, mode),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin", "import", "history"] }),
  });
}

const TERMINAL_STATUSES = new Set(["completed", "failed", "rolled_back"]);

/** Subscribes to the backend's SSE progress stream for one import job. */
export function useImportProgress(importLogId: number | null) {
  const [log, setLog] = useState<ImportLog | null>(null);
  const queryClient = useQueryClient();
  const sourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!importLogId) return;

    const source = new EventSource(adminImportService.streamUrl(importLogId), { withCredentials: true });
    sourceRef.current = source;

    source.addEventListener("progress", (event) => {
      const parsed = JSON.parse((event as MessageEvent).data) as ImportLog;
      setLog(parsed);
      if (TERMINAL_STATUSES.has(parsed.status)) {
        queryClient.invalidateQueries({ queryKey: ["admin", "import", "history"] });
        queryClient.invalidateQueries({ queryKey: ["admin", "books"] });
        source.close();
      }
    });
    source.addEventListener("error", () => source.close());

    return () => source.close();
  }, [importLogId, queryClient]);

  return log;
}
