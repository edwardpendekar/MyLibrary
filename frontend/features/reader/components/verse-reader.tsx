"use client";

import { useEffect, useState } from "react";

import { useChapterContent } from "@/hooks/use-books";
import { useSaveLastPosition } from "@/hooks/use-bookmarks";
import { useAuthStore } from "@/store/auth-store";
import { useReaderPreferencesStore } from "@/store/reader-preferences-store";
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
  const { data, isPending } = useChapterContent(bookId, chapterNumber);
  const { fontSize, translation } = useReaderPreferencesStore();
  const user = useAuthStore((s) => s.user);
  const saveLastPosition = useSaveLastPosition();
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
          {data.chapter.title_en ?? `Chapter ${chapterNumber}`}
        </h1>
        {versesBySection.map(({ section, verses }) => (
          <section key={section?.id ?? "no-section"} className="mb-8">
            {section?.title_en && <h2 className="mb-3 text-lg font-semibold text-primary">{section.title_en}</h2>}
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
                />
              ))}
            </div>
          </section>
        ))}
      </div>

      <AddNoteDialog bookId={bookId} verseId={noteVerseId} onOpenChange={(open) => !open && setNoteVerseId(null)} />
    </div>
  );
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
