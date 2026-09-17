"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
  useTranslateChapter,
} from "@/hooks/admin/use-admin-import";

/**
 * Translates one chapter at a time: the admin pastes a chapter's English
 * title + body, which becomes a single Gemini call (see
 * backend/scripts/translate_book.py) instead of splitting a whole-book
 * document into many sequential calls — that approach turned out to be slow
 * and fragile against transient API errors on long books. The result is just
 * an ImportPreview like any CSV upload, reviewed/committed through the same
 * ImportPreviewCard.
 */
export function TranslateUploader() {
  const t = useTranslations("admin.import_");
  const { data: books } = useAdminBooks();
  const [bookId, setBookId] = useState<string>("");
  const [chapterNumber, setChapterNumber] = useState<string>("");
  const [titleEn, setTitleEn] = useState("");
  const [bodyEn, setBodyEn] = useState("");
  const [jobId, setJobId] = useState<number | null>(null);
  const [mode, setMode] = useState<"insert" | "upsert">("insert");
  const [committedId, setCommittedId] = useState<number | null>(null);

  const translate = useTranslateChapter();
  const translateStatus = useImportProgress(jobId);
  const preview = useImportPreview(translateStatus?.status === "ready" ? jobId : null);
  const commit = useCommitImport();
  const commitProgress = useImportProgress(committedId);

  const isBusy = translate.isPending || (!!jobId && !preview.data && translateStatus?.status !== "failed");

  function handleSubmit() {
    const chapter = Number(chapterNumber);
    if (!bookId || !chapter || chapter < 1) {
      toast.error(t("translateMissingFields"));
      return;
    }
    if (!bodyEn.trim()) {
      toast.error(t("translateMissingFields"));
      return;
    }
    setJobId(null);
    setCommittedId(null);
    translate.mutate(
      { bookId: Number(bookId), chapterNumber: chapter, titleEn, bodyEn },
      {
        onSuccess: ({ data }) => setJobId(data.id),
        onError: (err) => toast.error(err instanceof Error ? err.message : t("translateFailed")),
      }
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="space-y-4 pt-6">
          <p className="text-sm text-muted-foreground">{t("translateHint")}</p>

          <div className="grid gap-4 sm:grid-cols-[2fr_1fr]">
            <div className="space-y-1.5">
              <Label>{t("translateSelectBook")}</Label>
              <Select value={bookId} onValueChange={(v) => setBookId(v ?? "")} disabled={isBusy}>
                <SelectTrigger className="w-full">
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
            </div>
            <div className="space-y-1.5">
              <Label>{t("translateChapterNumber")}</Label>
              <Input
                type="number"
                min={1}
                value={chapterNumber}
                onChange={(e) => setChapterNumber(e.target.value)}
                disabled={isBusy}
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>{t("translateChapterTitle")}</Label>
            <Input value={titleEn} onChange={(e) => setTitleEn(e.target.value)} disabled={isBusy} />
          </div>

          <div className="space-y-1.5">
            <Label>{t("translateChapterBody")}</Label>
            <Textarea
              rows={12}
              value={bodyEn}
              onChange={(e) => setBodyEn(e.target.value)}
              placeholder={t("translateChapterBodyPlaceholder")}
              disabled={isBusy}
            />
          </div>

          <Button onClick={handleSubmit} disabled={isBusy}>
            {isBusy ? t("translating") : t("translateSubmit")}
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
