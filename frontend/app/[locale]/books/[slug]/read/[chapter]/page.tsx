import { notFound } from "next/navigation";

import { apiServer } from "@/lib/api-server";
import { ApiError } from "@/lib/api-error";
import { ChapterList } from "@/features/reader/components/chapter-list";
import { VerseReader } from "@/features/reader/components/verse-reader";
import { RightPanel } from "@/features/reader/components/right-panel";
import type { Book } from "@/types/api";

export default async function ReaderPage({
  params,
}: PageProps<"/[locale]/books/[slug]/read/[chapter]">) {
  const { slug, chapter } = await params;
  const chapterNumber = Number(chapter);

  let book: Book;
  try {
    const res = await apiServer.get<Book>(`/api/v1/books/by-slug/${slug}`);
    book = res.data;
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound();
    throw err;
  }

  if (!Number.isFinite(chapterNumber) || chapterNumber < 1) notFound();

  return (
    <div className="grid h-[calc(100vh-3.5rem)] grid-cols-1 md:grid-cols-[200px_1fr_280px]">
      <aside className="hidden border-r md:block">
        <ChapterList bookId={book.id} bookSlug={book.slug} activeChapter={chapterNumber} />
      </aside>

      <main className="min-w-0">
        <VerseReader bookId={book.id} bookSlug={book.slug} chapterNumber={chapterNumber} />
      </main>

      <aside className="hidden border-l lg:block">
        <RightPanel bookId={book.id} bookSlug={book.slug} />
      </aside>
    </div>
  );
}
