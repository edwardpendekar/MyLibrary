"use client";

import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { useImportHistory } from "@/hooks/admin/use-admin-import";

const STATUS_VARIANT: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  completed: "default",
  failed: "destructive",
  rolled_back: "destructive",
  importing: "secondary",
  ready: "outline",
};

export function ImportHistory() {
  const { data: logs, isPending } = useImportHistory();

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>File</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Rows</TableHead>
            <TableHead>Inserted</TableHead>
            <TableHead>Updated</TableHead>
            <TableHead>Skipped</TableHead>
            <TableHead>Created</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isPending ? (
            Array.from({ length: 4 }).map((_, i) => (
              <TableRow key={i}>
                <TableCell colSpan={7}>
                  <Skeleton className="h-6 w-full" />
                </TableCell>
              </TableRow>
            ))
          ) : logs && logs.length > 0 ? (
            logs.map((log) => (
              <TableRow key={log.id}>
                <TableCell>{log.filename}</TableCell>
                <TableCell>
                  <Badge variant={STATUS_VARIANT[log.status] ?? "outline"}>{log.status}</Badge>
                </TableCell>
                <TableCell>
                  {log.processed_rows}/{log.total_rows}
                </TableCell>
                <TableCell>{log.verses_inserted}</TableCell>
                <TableCell>{log.verses_updated}</TableCell>
                <TableCell>{log.verses_skipped}</TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {new Date(log.created_at).toLocaleString()}
                </TableCell>
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell colSpan={7} className="text-center text-muted-foreground">
                No imports yet.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
}
