"use client";

import { useTranslations } from "next-intl";

import { Link } from "@/i18n/navigation";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { BookCard } from "@/features/books/components/book-card";
import { useSearchBooks, useSearchVerses } from "@/hooks/use-search";
import { sanitizeSnippet } from "@/lib/sanitize-snippet";

export function SearchResults({ query }: { query: string }) {
  const t = useTranslations("search");

  if (query.trim().length < 2) {
    return null;
  }

  return (
    <Tabs defaultValue="verses">
      <TabsList>
        <TabsTrigger value="verses">{t("versesTab")}</TabsTrigger>
        <TabsTrigger value="books">{t("booksTab")}</TabsTrigger>
      </TabsList>
      <TabsContent value="verses" className="mt-4">
        <VerseResults query={query} />
      </TabsContent>
      <TabsContent value="books" className="mt-4">
        <BookResults query={query} />
      </TabsContent>
    </Tabs>
  );
}

function VerseResults({ query }: { query: string }) {
  const t = useTranslations("search");
  const { data: hits, isPending } = useSearchVerses(query);

  if (isPending) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    );
  }

  if (!hits || hits.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("noResults", { query })}</p>;
  }

  return (
    <ul className="space-y-2">
      {hits.map((hit) => (
        <li key={hit.verse_id} className="rounded-lg border p-3">
          <Link
            href={`/books/${hit.book_slug}/read/${hit.chapter_number}#v${hit.verse_number}`}
            className="text-sm font-medium hover:underline"
          >
            {hit.book_title} {hit.chapter_number}:{hit.verse_number}
          </Link>
          <p
            className="mt-1 text-sm text-muted-foreground [&_b]:text-foreground [&_b]:font-semibold"
            dangerouslySetInnerHTML={{ __html: sanitizeSnippet(hit.snippet) }}
          />
        </li>
      ))}
    </ul>
  );
}

function BookResults({ query }: { query: string }) {
  const t = useTranslations("search");
  const { data: books, isPending } = useSearchBooks(query);

  if (isPending) {
    return (
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 md:grid-cols-6">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="aspect-2/3 w-full rounded-md" />
        ))}
      </div>
    );
  }

  if (!books || books.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("noResults", { query })}</p>;
  }

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 md:grid-cols-6">
      {books.map((book) => (
        <BookCard key={book.id} book={book} />
      ))}
    </div>
  );
}
