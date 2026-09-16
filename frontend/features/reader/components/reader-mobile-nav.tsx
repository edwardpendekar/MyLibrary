"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { BookmarkIcon, ListIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { ChapterList } from "@/features/reader/components/chapter-list";
import { RightPanel } from "@/features/reader/components/right-panel";

/**
 * Below the md/lg breakpoints where the reader's side columns are hidden
 * (see ReaderPage), the chapter list and bookmarks/notes/search panel are
 * still reachable — as slide-over sheets triggered from a small toolbar —
 * rather than simply disappearing on phones/small tablets.
 */
export function ReaderMobileNav({
  bookId,
  bookSlug,
  activeChapter,
}: {
  bookId: number;
  bookSlug: string;
  activeChapter: number;
}) {
  const t = useTranslations("reader");
  const [chaptersOpen, setChaptersOpen] = useState(false);
  const [panelOpen, setPanelOpen] = useState(false);

  return (
    <div className="flex items-center gap-1 border-b bg-background px-2 py-1.5 lg:hidden">
      <Sheet open={chaptersOpen} onOpenChange={setChaptersOpen}>
        <Button
          variant="ghost"
          size="sm"
          className="md:hidden"
          onClick={() => setChaptersOpen(true)}
        >
          <ListIcon className="size-4" />
          {t("chapters")}
        </Button>
        <SheetContent side="left" className="w-72 p-0">
          <SheetTitle className="sr-only">{t("chapters")}</SheetTitle>
          <ChapterList
            bookId={bookId}
            bookSlug={bookSlug}
            activeChapter={activeChapter}
            onNavigate={() => setChaptersOpen(false)}
          />
        </SheetContent>
      </Sheet>

      <Sheet open={panelOpen} onOpenChange={setPanelOpen}>
        <Button variant="ghost" size="sm" onClick={() => setPanelOpen(true)}>
          <BookmarkIcon className="size-4" />
          {t("bookmarks")}/{t("notes")}
        </Button>
        <SheetContent side="right" className="w-80 p-0">
          <SheetTitle className="sr-only">{t("bookmarks")}</SheetTitle>
          <RightPanel bookId={bookId} bookSlug={bookSlug} />
        </SheetContent>
      </Sheet>
    </div>
  );
}
