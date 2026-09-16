"use client";

import { useRef, useState } from "react";
import { Upload } from "lucide-react";
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
import { useCommitImport, useImportProgress, useUploadImportFile } from "@/hooks/admin/use-admin-import";
import type { ImportPreview } from "@/types/api";

export function ImportUploader() {
  const t = useTranslations("admin.import_");
  const inputRef = useRef<HTMLInputElement | null>(null);
  const upload = useUploadImportFile();
  const commit = useCommitImport();
  const [preview, setPreview] = useState<ImportPreview | null>(null);
  const [mode, setMode] = useState<"insert" | "upsert">("insert");
  const [committedId, setCommittedId] = useState<number | null>(null);
  const progress = useImportProgress(committedId);

  function handleFile(file: File) {
    setPreview(null);
    setCommittedId(null);
    upload.mutate(file, { onSuccess: ({ data }) => setPreview(data) });
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="flex flex-col items-center gap-3 py-10 text-center">
          <input
            ref={inputRef}
            type="file"
            accept=".xlsx,.xls,.csv"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) handleFile(file);
              e.target.value = "";
            }}
          />
          <Upload className="size-8 text-muted-foreground" />
          <div>
            <p className="font-medium">{t("uploadTitle")}</p>
            <p className="text-sm text-muted-foreground">{t("uploadColumnsHint")}</p>
          </div>
          <Button onClick={() => inputRef.current?.click()} disabled={upload.isPending}>
            {upload.isPending ? t("validating") : t("chooseFile")}
          </Button>
        </CardContent>
      </Card>

      {preview && (
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
              <RadioGroup value={mode} onValueChange={(v) => setMode(v as "insert" | "upsert")}>
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

            <Button
              onClick={() =>
                commit.mutate(
                  { id: preview.import_log_id, mode },
                  { onSuccess: () => setCommittedId(preview.import_log_id) }
                )
              }
              disabled={commit.isPending || !!committedId}
            >
              {t("commit")}
            </Button>
          </CardContent>
        </Card>
      )}

      {committedId && progress && (
        <Card>
          <CardContent className="space-y-2 pt-6 text-sm">
            <p className="font-medium">
              {t("statusLabel")}: {t(`status.${progress.status}`)}
            </p>
            <p>
              {t("processed")} {progress.processed_rows} / {progress.total_rows}
            </p>
            <p>
              {t("booksCreated")} +{progress.books_created} · {t("chaptersCreated")} +{progress.chapters_created} ·{" "}
              {t("sectionsCreated")} +{progress.sections_created}
            </p>
            <p>
              {t("verses")} {t("inserted")} {progress.verses_inserted} · {t("updated")} {progress.verses_updated} ·{" "}
              {t("skipped")} {progress.verses_skipped}
            </p>
            {progress.error_message && <p className="text-destructive">{progress.error_message}</p>}
          </CardContent>
        </Card>
      )}
    </div>
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
