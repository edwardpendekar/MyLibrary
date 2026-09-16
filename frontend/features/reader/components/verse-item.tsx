"use client";

import { useTranslations } from "next-intl";
import { Bookmark, Copy, Highlighter, MessageSquarePlus, Share2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { useHighlightsStore } from "@/store/highlights-store";
import { useAuthStore } from "@/store/auth-store";
import { useCreateBookmark } from "@/hooks/use-bookmarks";
import { cn } from "@/lib/utils";
import type { ReaderTranslation } from "@/store/reader-preferences-store";
import type { Verse } from "@/types/api";

export function VerseItem({
  verse,
  bookId,
  bookSlug,
  chapterNumber,
  translation,
  fontSize,
  onAddNote,
  highlightedVerseIds,
  onToggleHighlight,
}: {
  verse: Verse;
  bookId: number;
  bookSlug: string;
  chapterNumber: number;
  translation: ReaderTranslation;
  fontSize: number;
  onAddNote: (verseId: number) => void;
  /** Present (possibly empty) only when logged in — account-synced highlights. */
  highlightedVerseIds?: Set<number>;
  onToggleHighlight?: (verseId: number, currentlyHighlighted: boolean) => void;
}) {
  const t = useTranslations("reader");
  const user = useAuthStore((s) => s.user);
  const localHighlighted = useHighlightsStore((s) => s.isHighlighted(verse.id));
  const toggleLocalHighlight = useHighlightsStore((s) => s.toggle);
  const createBookmark = useCreateBookmark(bookId);

  // Guests keep the device-local (localStorage) highlight set; logged-in users
  // get the account-synced set fetched once for the whole book by the parent.
  const isHighlighted = user ? (highlightedVerseIds?.has(verse.id) ?? false) : localHighlighted;
  function toggleHighlight() {
    if (user) {
      onToggleHighlight?.(verse.id, isHighlighted);
    } else {
      toggleLocalHighlight(verse.id);
    }
  }

  const text =
    translation === "en" ? verse.text_en : translation === "id" ? verse.text_id : null;

  async function copyVerse() {
    const parts = [verse.text_en, verse.text_id].filter(Boolean);
    await navigator.clipboard.writeText(`${chapterNumber}:${verse.number} ${parts.join(" / ")}`);
    toast.success(t("copied"));
  }

  async function shareVerse() {
    const url = `${window.location.origin}/books/${bookSlug}/read/${chapterNumber}#v${verse.number}`;
    if (navigator.share) {
      await navigator.share({ url }).catch(() => {});
    } else {
      await navigator.clipboard.writeText(url);
      toast.success(t("copied"));
    }
  }

  return (
    <div
      id={`v${verse.number}`}
      className={cn(
        "group scroll-mt-24 rounded-md px-2 py-1.5 -mx-2 transition-colors",
        isHighlighted && "bg-yellow-100 dark:bg-yellow-900/30"
      )}
    >
      <div className="flex items-start gap-2">
        <sup className="mt-1.5 text-xs font-semibold text-muted-foreground">{verse.number}</sup>
        <div className="flex-1 space-y-1" style={{ fontSize }}>
          {translation === "both" ? (
            <>
              {verse.text_en && (
                <p lang="en" className="leading-relaxed">
                  {verse.text_en}
                </p>
              )}
              {verse.text_id && (
                <p lang="id" className="leading-relaxed text-muted-foreground">
                  {verse.text_id}
                </p>
              )}
            </>
          ) : (
            text && (
              <p lang={translation} className="leading-relaxed">
                {text}
              </p>
            )
          )}
        </div>
      </div>

      <div className="ml-6 flex gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
        <Button variant="ghost" size="icon-xs" aria-label={t("copy")} onClick={copyVerse}>
          <Copy className="size-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon-xs"
          aria-label={t("highlight")}
          onClick={toggleHighlight}
        >
          <Highlighter className={cn("size-3.5", isHighlighted && "text-yellow-600")} />
        </Button>
        {user && (
          <>
            <Button
              variant="ghost"
              size="icon-xs"
              aria-label={t("addBookmark")}
              onClick={() =>
                createBookmark.mutate(
                  { book_id: bookId, chapter_id: undefined, verse_id: verse.id },
                  { onSuccess: () => toast.success(t("addBookmark")) }
                )
              }
            >
              <Bookmark className="size-3.5" />
            </Button>
            <Button variant="ghost" size="icon-xs" aria-label={t("addNote")} onClick={() => onAddNote(verse.id)}>
              <MessageSquarePlus className="size-3.5" />
            </Button>
          </>
        )}
        <Button variant="ghost" size="icon-xs" aria-label={t("share")} onClick={shareVerse}>
          <Share2 className="size-3.5" />
        </Button>
      </div>
    </div>
  );
}
