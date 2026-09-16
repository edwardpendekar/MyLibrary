"use client";

import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";

import { Link } from "@/i18n/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { useAdminBooks, useCreateBook, useDeleteBook } from "@/hooks/admin/use-admin-books";
import { BookForm } from "@/features/admin/components/book-form";
import { useDebounce } from "@/hooks/use-debounce";

export default function AdminBooksPage() {
  const tAdmin = useTranslations("admin");
  const t = useTranslations("admin.booksPage");
  const [query, setQuery] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const debounced = useDebounce(query);
  const { data: books, isPending } = useAdminBooks(debounced);
  const createBook = useCreateBook();
  const deleteBook = useDeleteBook();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{tAdmin("books")}</h1>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="size-4" />
          {t("newBook")}
        </Button>
      </div>

      <Input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder={t("searchPlaceholder")}
        className="max-w-sm"
      />

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("colTitle")}</TableHead>
              <TableHead>{t("colStatus")}</TableHead>
              <TableHead>{t("colChapters")}</TableHead>
              <TableHead>{t("colVerses")}</TableHead>
              <TableHead className="text-right">{t("colActions")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isPending ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell colSpan={5}>
                    <Skeleton className="h-6 w-full" />
                  </TableCell>
                </TableRow>
              ))
            ) : books && books.length > 0 ? (
              books.map((book) => (
                <TableRow key={book.id}>
                  <TableCell>
                    <Link href={`/admin/books/${book.id}`} className="font-medium hover:underline">
                      {book.title}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Badge variant={book.status === "published" ? "default" : "secondary"}>{book.status}</Badge>
                  </TableCell>
                  <TableCell>{book.chapters_count}</TableCell>
                  <TableCell>{book.verses_count}</TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => {
                        if (confirm(t("confirmDelete", { title: book.title }))) deleteBook.mutate(book.id);
                      }}
                    >
                      <Trash2 className="size-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-muted-foreground">
                  {t("noBooks")}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("newBook")}</DialogTitle>
          </DialogHeader>
          <BookForm
            onSubmit={(input) =>
              createBook.mutate(input, { onSuccess: () => setCreateOpen(false) })
            }
            isSubmitting={createBook.isPending}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}
