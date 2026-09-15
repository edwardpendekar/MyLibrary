"use client";

import { useTranslations } from "next-intl";

import { usePopularBooks } from "@/hooks/use-books";
import { BookCard } from "@/features/books/components/book-card";
import { Skeleton } from "@/components/ui/skeleton";

export function PopularRail() {
  const t = useTranslations("home");
  const { data: books, isPending } = usePopularBooks(10);

  if (isPending) {
    return (
      <div className="flex gap-4 overflow-x-auto pb-2">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="aspect-2/3 w-32 shrink-0 rounded-md" />
        ))}
      </div>
    );
  }

  if (!books || books.length === 0) return null;

  return (
    <section className="space-y-3">
      <h2 className="text-lg font-semibold">{t("popular")}</h2>
      <div className="flex gap-4 overflow-x-auto pb-2">
        {books.map((book) => (
          <div key={book.id} className="w-32 shrink-0 sm:w-36">
            <BookCard book={book} />
          </div>
        ))}
      </div>
    </section>
  );
}
