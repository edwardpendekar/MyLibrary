import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { BookOpen, Download, FileText } from "lucide-react";

import { apiServer } from "@/lib/api-server";
import { Link } from "@/i18n/navigation";
import { buttonVariants } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { BookCover } from "@/features/books/components/book-cover";
import { FavoriteButton } from "@/features/books/components/favorite-button";
import { ShareButton } from "@/features/books/components/share-button";
import { ApiError } from "@/lib/api-error";
import type { Book } from "@/types/api";

export default async function BookDetailPage({
  params,
}: PageProps<"/[locale]/books/[slug]">) {
  const { slug } = await params;
  const t = await getTranslations("book");

  let book: Book;
  try {
    const res = await apiServer.get<Book>(`/api/v1/books/by-slug/${slug}`);
    book = res.data;
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound();
    throw err;
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <div className="grid gap-8 sm:grid-cols-[240px_1fr]">
        <BookCover src={book.cover_url} alt={book.title} sizes="240px" priority />

        <div className="space-y-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{book.title}</h1>
            {book.author && <p className="text-muted-foreground mt-1">{t("by", { author: book.author })}</p>}
          </div>

          <div className="flex flex-wrap gap-1.5">
            {book.category && <Badge variant="secondary">{book.category.name_en}</Badge>}
            {book.language && <Badge variant="outline">{book.language.name}</Badge>}
            {book.year && <Badge variant="outline">{book.year}</Badge>}
          </div>

          <p className="text-sm text-muted-foreground">
            {book.chapters_count} {t("chapters")} · {book.verses_count} {t("verses")}
          </p>

          {book.description && (
            <div>
              <h2 className="font-medium">{t("description")}</h2>
              <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{book.description}</p>
            </div>
          )}

          {/* Read/Open PDF/Download are navigation, not in-page actions, so
              they get button styling on a real <Link>/<a> rather than being
              wrapped in the Button component — see SiteHeader for why. */}
          <div className="flex flex-wrap gap-2 pt-2">
            <Link href={`/books/${book.slug}/read/1`} className={buttonVariants()}>
              <BookOpen className="size-4" />
              {t("read")}
            </Link>
            {book.has_pdf && book.pdf_url && (
              <>
                <Link href={`/books/${book.slug}/pdf`} className={buttonVariants({ variant: "outline" })}>
                  <FileText className="size-4" />
                  {t("openPdf")}
                </Link>
                <a href={book.pdf_url} download className={buttonVariants({ variant: "ghost" })}>
                  <Download className="size-4" />
                  {t("downloadPdf")}
                </a>
              </>
            )}
            <FavoriteButton bookId={book.id} isFavorite={!!book.is_favorite} />
            <ShareButton title={book.title} path={`/books/${book.slug}`} />
          </div>
        </div>
      </div>
    </div>
  );
}
