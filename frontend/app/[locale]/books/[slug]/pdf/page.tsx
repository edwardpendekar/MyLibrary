import { notFound } from "next/navigation";

import { apiServer } from "@/lib/api-server";
import { ApiError } from "@/lib/api-error";
import { PdfViewer } from "@/features/reader/components/pdf-viewer";
import type { Book } from "@/types/api";

export default async function BookPdfPage({
  params,
}: PageProps<"/[locale]/books/[slug]/pdf">) {
  const { slug } = await params;

  let book: Book;
  try {
    const res = await apiServer.get<Book>(`/api/v1/books/by-slug/${slug}`);
    book = res.data;
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound();
    throw err;
  }

  if (!book.has_pdf || !book.pdf_url) notFound();

  return <PdfViewer bookId={book.id} url={book.pdf_url} />;
}
