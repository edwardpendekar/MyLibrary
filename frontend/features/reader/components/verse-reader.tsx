"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";

import { useChapterContent } from "@/hooks/use-books";
import { useSaveLastPosition } from "@/hooks/use-bookmarks";
import { useHighlights, useToggleHighlight } from "@/hooks/use-highlights";
import { useAuthStore } from "@/store/auth-store";
import { useReaderPreferencesStore, type ReaderTranslation } from "@/store/reader-preferences-store";
import { ReaderToolbar } from "@/features/reader/components/reader-toolbar";
import { VerseItem } from "@/features/reader/components/verse-item";
import { AddNoteDialog } from "@/features/reader/components/add-note-dialog";
import { Skeleton } from "@/components/ui/skeleton";

export function VerseReader({
  bookId,
  bookSlug,
  chapterNumber,
}: {
  bookId: number;
  bookSlug: string;
  chapterNumber: number;
}) {
  const t = useTranslations("reader");
  const { data, isPending } = useChapterContent(bookId, chapterNumber);
  const { fontSize, translation } = useReaderPreferencesStore();
  const user = useAuthStore((s) => s.user);
  const saveLastPosition = useSaveLastPosition();
  const { data: highlightedVerseIds } = useHighlights(bookId, !!user);
  const toggleHighlight = useToggleHighlight(bookId);
  const [noteVerseId, setNoteVerseId] = useState<number | null>(null);

  // "Remember last page": recorded at chapter granularity whenever a logged-in
  // reader opens a chapter. Precise in-chapter scroll position is intentionally
  // out of scope here to avoid a write on every scroll tick.
  useEffect(() => {
    if (!user || !data) return;
    saveLastPosition.mutate({ book_id: bookId, chapter_id: data.chapter.id });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, bookId, data?.chapter.id]);

  if (isPending) {
    return (
      <div className="space-y-3 p-6">
        {Array.from({ length: 8 }).map((_, i) => (
          <Skeleton key={i} className="h-5 w-full" />
        ))}
      </div>
    );
  }

  if (!data) return null;

  const versesBySection = groupBySection(data.verses, data.sections);

  return (
    <div className="flex h-full flex-col">
      <ReaderToolbar />
      <div className="flex-1 overflow-y-auto px-4 py-6 sm:px-8">
        <h1 className="mb-6 text-2xl font-bold">
          {pickLocalizedTitle(data.chapter.title_en, data.chapter.title_id, translation) ??
            t("chapterNumber", { number: chapterNumber })}
        </h1>
        {versesBySection.map(({ section, verses }) => {
          const sectionTitle = pickLocalizedTitle(section?.title_en, section?.title_id, translation);
          return (
            <section key={section?.id ?? "no-section"} className="mb-8">
              {sectionTitle && <h2 className="mb-3 text-lg font-semibold text-primary">{sectionTitle}</h2>}
              <div className="space-y-1">
                {verses.map((verse) => (
                  <VerseItem
                    key={verse.id}
                    verse={verse}
                    bookId={bookId}
                    bookSlug={bookSlug}
                    chapterNumber={chapterNumber}
                    translation={translation}
                    fontSize={fontSize}
                    onAddNote={setNoteVerseId}
                    highlightedVerseIds={user ? highlightedVerseIds : undefined}
                    onToggleHighlight={(verseId, highlighted) => toggleHighlight.mutate({ verseId, highlighted })}
                  />
                ))}
              </div>
            </section>
          );
        })}
      </div>

      <AddNoteDialog bookId={bookId} verseId={noteVerseId} onOpenChange={(open) => !open && setNoteVerseId(null)} />
    </div>
  );
}

/**
 * Chapter/section titles follow the same EN/ID/both reading preference as
 * verse text — "both" prefers Indonesian first since that's this app's
 * primary audience, falling back to whichever language is actually present.
 */
function pickLocalizedTitle(
  titleEn: string | null | undefined,
  titleId: string | null | undefined,
  translation: ReaderTranslation
): string | null {
  if (translation === "en") return titleEn || titleId || null;
  return titleId || titleEn || null;
}

function groupBySection<
  V extends { section_id?: number | null },
  S extends { id: number },
>(verses: V[], sections: S[]) {
  const sectionById = new Map(sections.map((s) => [s.id, s]));
  const groups: { section: S | null; verses: V[] }[] = [];

  for (const verse of verses) {
    const section = verse.section_id != null ? (sectionById.get(verse.section_id) ?? null) : null;
    const last = groups[groups.length - 1];
    if (last && last.section?.id === section?.id) {
      last.verses.push(verse);
    } else {
      groups.push({ section, verses: [verse] });
    }
  }
  return groups;
}
