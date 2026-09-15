"use client";

import { use } from "react";
import { useRouter } from "@/i18n/navigation";

import { BookForm } from "@/features/admin/components/book-form";
import { FileUploadCard } from "@/features/admin/components/file-upload-card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useAdminBook,
  useUpdateBook,
  useUploadCover,
  useUploadPdf,
} from "@/hooks/admin/use-admin-books";

export default function AdminBookEditPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const bookId = Number(id);
  const router = useRouter();

  const { data: book, isPending } = useAdminBook(bookId);
  const updateBook = useUpdateBook(bookId);
  const uploadCover = useUploadCover(bookId);
  const uploadPdf = useUploadPdf(bookId);

  if (isPending) {
    return <Skeleton className="h-96 w-full max-w-2xl" />;
  }

  if (!book) {
    return <p className="text-muted-foreground">Book not found.</p>;
  }

  return (
    <div className="max-w-2xl space-y-8">
      <h1 className="text-2xl font-bold">{book.title}</h1>

      <div className="grid grid-cols-2 gap-4">
        <FileUploadCard
          title="Cover image"
          accept="image/jpeg,image/png,image/webp"
          currentUrl={book.cover_url}
          isImage
          isUploading={uploadCover.isPending}
          onSelect={(file) => uploadCover.mutate(file)}
        />
        <FileUploadCard
          title="PDF ebook"
          accept="application/pdf"
          currentUrl={book.pdf_url}
          isUploading={uploadPdf.isPending}
          onSelect={(file) => uploadPdf.mutate(file)}
        />
      </div>

      <BookForm
        book={book}
        isSubmitting={updateBook.isPending}
        onSubmit={(input) =>
          updateBook.mutate(input, { onSuccess: () => router.push("/admin/books") })
        }
      />
    </div>
  );
}
