"use client";

import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import type { ImportPreview } from "@/types/api";

/**
 * The validation-preview + mode-select + commit screen shared by both import
 * entry points (plain CSV upload and the AI-translate-from-document flow) —
 * once a file has produced an ImportPreview, reviewing and committing it
 * looks identical either way.
 */
export function ImportPreviewCard({
  preview,
  mode,
  onModeChange,
  onCommit,
  commitPending,
  committed,
}: {
  preview: ImportPreview;
  mode: "insert" | "upsert";
  onModeChange: (mode: "insert" | "upsert") => void;
  onCommit: () => void;
  commitPending: boolean;
  committed: boolean;
}) {
  const t = useTranslations("admin.import_");

  return (
    <Card>
      <CardContent className="space-y-4 pt-6">
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <Stat label={t("rowsTotal")} value={preview.total_rows} />
          <Stat label={t("distinctBooks")} value={preview.distinct_books} />
          <Stat label={t("distinctChapters")} value={preview.distinct_chapters} />
          <Stat label={t("errors")} value={preview.validation_errors?.length ?? 0} />
        </div>

        {(preview.sample_rows?.length ?? 0) > 0 && (
          <div className="overflow-x-auto rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("row")}</TableHead>
                  <TableHead>{t("book")}</TableHead>
                  <TableHead>{t("ch")}</TableHead>
                  <TableHead>{t("vs")}</TableHead>
                  <TableHead>text_en</TableHead>
                  <TableHead>title_en</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(preview.sample_rows ?? []).slice(0, 10).map((row) => (
                  <TableRow key={row.row_number}>
                    <TableCell>{row.row_number}</TableCell>
                    <TableCell>{row.book}</TableCell>
                    <TableCell>{row.chapter}</TableCell>
                    <TableCell>{row.verse}</TableCell>
                    <TableCell className="max-w-56 truncate">{row.text_en}</TableCell>
                    <TableCell>{row.title_en}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}

        {(preview.validation_errors?.length ?? 0) > 0 && (
          <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm">
            <p className="mb-1 font-medium text-destructive">
              {t("validationErrors")} ({preview.validation_errors?.length ?? 0})
            </p>
            <ul className="max-h-32 space-y-0.5 overflow-y-auto text-muted-foreground">
              {(preview.validation_errors ?? []).map((e) => (
                <li key={e.row_number}>
                  {t("row")} {e.row_number}: {e.message}
                </li>
              ))}
            </ul>
          </div>
        )}

        <div className="flex items-center gap-4">
          <RadioGroup value={mode} onValueChange={(v) => onModeChange(v as "insert" | "upsert")}>
            <label className="flex items-center gap-2 text-sm">
              <RadioGroupItem value="insert" />
              {t("modeInsert")}
            </label>
            <label className="flex items-center gap-2 text-sm">
              <RadioGroupItem value="upsert" />
              {t("modeUpsert")}
            </label>
          </RadioGroup>
        </div>

        <Button onClick={onCommit} disabled={commitPending || committed}>
          {t("commit")}
        </Button>
      </CardContent>
    </Card>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="text-xl font-bold">{value}</p>
    </div>
  );
}
