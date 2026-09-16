"use client";

import { useTranslations } from "next-intl";

import { Link } from "@/i18n/navigation";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { useChapters } from "@/hooks/use-books";
import { cn } from "@/lib/utils";

export function ChapterList({
  bookId,
  bookSlug,
  activeChapter,
  onNavigate,
}: {
  bookId: number;
  bookSlug: string;
  activeChapter: number;
  /** Fired when a chapter link is clicked — lets the mobile sheet close itself. */
  onNavigate?: () => void;
}) {
  const t = useTranslations("reader");
  const { data: chapters, isPending } = useChapters(bookId);

  if (isPending) {
    return (
      <div className="space-y-1.5 p-3">
        {Array.from({ length: 10 }).map((_, i) => (
          <Skeleton key={i} className="h-6 w-full" />
        ))}
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      <h2 className="border-b px-3 py-2 text-sm font-semibold">{t("chapters")}</h2>
      <ScrollArea className="flex-1">
        <nav className="grid grid-cols-4 gap-1 p-3 sm:grid-cols-5">
          {chapters?.map((chapter) => (
            <Link
              key={chapter.id}
              href={`/books/${bookSlug}/read/${chapter.number}`}
              onClick={onNavigate}
              className={cn(
                "flex h-9 items-center justify-center rounded-md text-sm hover:bg-muted",
                chapter.number === activeChapter && "bg-primary text-primary-foreground hover:bg-primary/90"
              )}
            >
              {chapter.number}
            </Link>
          ))}
        </nav>
      </ScrollArea>
    </div>
  );
}
