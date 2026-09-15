"use client";

import { useTranslations } from "next-intl";

import { useFavorites } from "@/hooks/use-favorites";
import { BookCard } from "@/features/books/components/book-card";
import { Skeleton } from "@/components/ui/skeleton";

export default function FavoritesPage() {
  const t = useTranslations("nav");
  const tHome = useTranslations("home");
  const { data: books, isPending } = useFavorites();

  return (
    <div className="mx-auto max-w-7xl space-y-6 px-4 py-8">
      <h1 className="text-2xl font-bold">{t("favorites")}</h1>

      {isPending ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="aspect-2/3 w-full rounded-md" />
          ))}
        </div>
      ) : books && books.length > 0 ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
          {books.map((book) => (
            <BookCard key={book.id} book={book} />
          ))}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">{tHome("noResults")}</p>
      )}
    </div>
  );
}
