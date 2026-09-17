"use client";

import { useRef, useState } from "react";
import { Upload } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { ImportPreviewCard } from "@/features/admin/components/import-preview-card";
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
        <ImportPreviewCard
          preview={preview}
          mode={mode}
          onModeChange={setMode}
          commitPending={commit.isPending}
          committed={!!committedId}
          onCommit={() =>
            commit.mutate(
              { id: preview.import_log_id, mode },
              { onSuccess: () => setCommittedId(preview.import_log_id) }
            )
          }
        />
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
