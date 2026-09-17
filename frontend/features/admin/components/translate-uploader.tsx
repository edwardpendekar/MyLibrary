"use client";

import { useRef, useState } from "react";
import { FileText } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ImportPreviewCard } from "@/features/admin/components/import-preview-card";
import { useAdminBooks } from "@/hooks/admin/use-admin-books";
import {
  useCommitImport,
  useImportPreview,
  useImportProgress,
  useTranslateBook,
} from "@/hooks/admin/use-admin-import";

/**
 * Uploads a whole English manuscript (.docx/.txt) for an existing book and
 * has the backend translate + split it into TB2-style Indonesian verses via
 * Gemini (see backend/scripts/translate_book.py). Once that background job
 * flips to "ready", the result is just an ImportPreview like any CSV upload —
 * reviewed and committed through the exact same ImportPreviewCard.
 */
export function TranslateUploader() {
  const t = useTranslations("admin.import_");
  const inputRef = useRef<HTMLInputElement | null>(null);
  const { data: books } = useAdminBooks();
  const [bookId, setBookId] = useState<string>("");
  const [jobId, setJobId] = useState<number | null>(null);
  const [mode, setMode] = useState<"insert" | "upsert">("insert");
  const [committedId, setCommittedId] = useState<number | null>(null);

  const translate = useTranslateBook();
  const translateStatus = useImportProgress(jobId);
  const preview = useImportPreview(translateStatus?.status === "ready" ? jobId : null);
  const commit = useCommitImport();
  const commitProgress = useImportProgress(committedId);

  const isBusy = translate.isPending || (!!jobId && !preview.data && translateStatus?.status !== "failed");

  function handleFile(file: File) {
    if (!bookId) {
      toast.error(t("translateSelectBookFirst"));
      return;
    }
    setJobId(null);
    setCommittedId(null);
    translate.mutate(
      { file, bookId: Number(bookId) },
      {
        onSuccess: ({ data }) => setJobId(data.id),
        onError: (err) => toast.error(err instanceof Error ? err.message : t("translateFailed")),
      }
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="flex flex-col items-center gap-3 py-10 text-center">
          <input
            ref={inputRef}
            type="file"
            accept=".docx,.txt"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) handleFile(file);
              e.target.value = "";
            }}
          />
          <FileText className="size-8 text-muted-foreground" />
          <div>
            <p className="font-medium">{t("translateTitle")}</p>
            <p className="text-sm text-muted-foreground">{t("translateHint")}</p>
          </div>
          <Select value={bookId} onValueChange={(v) => setBookId(v ?? "")}>
            <SelectTrigger className="w-72">
              <SelectValue placeholder={t("translateSelectBook")} />
            </SelectTrigger>
            <SelectContent>
              {books?.map((b) => (
                <SelectItem key={b.id} value={String(b.id)}>
                  {b.title}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button onClick={() => inputRef.current?.click()} disabled={!bookId || isBusy}>
            {isBusy ? t("translating") : t("chooseFile")}
          </Button>
        </CardContent>
      </Card>

      {jobId && translateStatus && !preview.data && (
        <Card>
          <CardContent className="space-y-2 pt-6 text-sm">
            <p className="font-medium">
              {t("statusLabel")}: {t(`status.${translateStatus.status}`)}
            </p>
            {translateStatus.error_message && <p className="text-destructive">{translateStatus.error_message}</p>}
          </CardContent>
        </Card>
      )}

      {preview.data && (
        <ImportPreviewCard
          preview={preview.data}
          mode={mode}
          onModeChange={setMode}
          commitPending={commit.isPending}
          committed={!!committedId}
          onCommit={() => {
            const id = preview.data!.import_log_id;
            commit.mutate({ id, mode }, { onSuccess: () => setCommittedId(id) });
          }}
        />
      )}

      {committedId && commitProgress && (
        <Card>
          <CardContent className="space-y-2 pt-6 text-sm">
            <p className="font-medium">
              {t("statusLabel")}: {t(`status.${commitProgress.status}`)}
            </p>
            <p>
              {t("processed")} {commitProgress.processed_rows} / {commitProgress.total_rows}
            </p>
            <p>
              {t("booksCreated")} +{commitProgress.books_created} · {t("chaptersCreated")} +
              {commitProgress.chapters_created} · {t("sectionsCreated")} +{commitProgress.sections_created}
            </p>
            <p>
              {t("verses")} {t("inserted")} {commitProgress.verses_inserted} · {t("updated")}{" "}
              {commitProgress.verses_updated} · {t("skipped")} {commitProgress.verses_skipped}
            </p>
            {commitProgress.error_message && <p className="text-destructive">{commitProgress.error_message}</p>}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
