"use client";

import { useTranslations } from "next-intl";

import { useInfiniteBooks } from "@/hooks/use-books";
import { useIntersectionObserver } from "@/hooks/use-intersection-observer";
import { BookCard } from "@/features/books/components/book-card";
import { Skeleton } from "@/components/ui/skeleton";
import type { ListBooksParams } from "@/services/books.service";

export function BookGrid({ filters = {} }: { filters?: Omit<ListBooksParams, "cursor"> }) {
  const t = useTranslations("home");
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending, isError } =
    useInfiniteBooks(filters);

  const sentinelRef = useIntersectionObserver(() => {
    if (hasNextPage && !isFetchingNextPage) fetchNextPage();
  }, !!hasNextPage);

  // isPending (not isLoading): during SSR the query hasn't started fetching yet
  // at all (TanStack Query only kicks off the request client-side after mount),
  // so isLoading (which also requires isFetching) would be false and this would
  // otherwise fall through to the "no results" branch before the real fetch runs.
  if (isPending) {
    return (
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
        {Array.from({ length: 12 }).map((_, i) => (
          <Skeleton key={i} className="aspect-2/3 w-full rounded-md" />
        ))}
      </div>
    );
  }

  if (isError) {
    return <p className="text-muted-foreground text-sm">{t("noResults")}</p>;
  }

  const books = data?.pages.flatMap((page) => page.data) ?? [];

  if (books.length === 0) {
    return <p className="text-muted-foreground text-sm">{t("noResults")}</p>;
  }

  return (
    <div>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
        {books.map((book, i) => (
          <BookCard key={book.id} book={book} priority={i < 6} />
        ))}
      </div>
      <div ref={sentinelRef} className="h-10" />
      {isFetchingNextPage && (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 mt-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="aspect-2/3 w-full rounded-md" />
          ))}
        </div>
      )}
    </div>
  );
}
