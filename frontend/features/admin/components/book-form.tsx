"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslations } from "next-intl";
import { z } from "zod";

import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAdminLanguages, useAdminCategories } from "@/hooks/admin/use-admin-reference";
import type { UpsertBookInput } from "@/services/admin/books.service";
import type { Book } from "@/types/api";

const schema = z.object({
  title: z.string().min(1).max(255),
  author: z.string().max(255).optional(),
  description: z.string().optional(),
  language_id: z.string().optional(),
  category_id: z.string().optional(),
  year: z.string().optional(),
  isbn: z.string().optional(),
  status: z.enum(["draft", "published", "archived"]).optional(),
});

type FormValues = z.infer<typeof schema>;

export function BookForm({
  book,
  onSubmit,
  isSubmitting,
}: {
  book?: Book;
  onSubmit: (input: UpsertBookInput) => void;
  isSubmitting?: boolean;
}) {
  const t = useTranslations("admin.booksPage.form");
  const { data: languages } = useAdminLanguages();
  const { data: categories } = useAdminCategories();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: book?.title ?? "",
      author: book?.author ?? "",
      description: book?.description ?? "",
      language_id: book?.language?.id ? String(book.language.id) : undefined,
      category_id: book?.category?.id ? String(book.category.id) : undefined,
      year: book?.year ? String(book.year) : "",
      isbn: book?.isbn ?? "",
      status: book?.status ?? "draft",
    },
  });

  function handleSubmit(values: FormValues) {
    onSubmit({
      title: values.title,
      author: values.author || undefined,
      description: values.description || undefined,
      language_id: values.language_id ? Number(values.language_id) : undefined,
      category_id: values.category_id ? Number(values.category_id) : undefined,
      year: values.year ? Number(values.year) : undefined,
      isbn: values.isbn || undefined,
      status: values.status,
    });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="title"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("title")}</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="author"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("author")}</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("description")}</FormLabel>
              <FormControl>
                <Textarea rows={3} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="grid grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="language_id"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("language")}</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder={t("selectLanguage")} />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {languages?.map((l) => (
                      <SelectItem key={l.id} value={String(l.id)}>
                        {l.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="category_id"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("category")}</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder={t("selectCategory")} />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {categories?.map((c) => (
                      <SelectItem key={c.id} value={String(c.id)}>
                        {c.name_en}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormItem>
            )}
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="year"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("year")}</FormLabel>
                <FormControl>
                  <Input type="number" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="isbn"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("isbn")}</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        {book && (
          <FormField
            control={form.control}
            name="status"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("status")}</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl>
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="draft">{t("statusDraft")}</SelectItem>
                    <SelectItem value="published">{t("statusPublished")}</SelectItem>
                    <SelectItem value="archived">{t("statusArchived")}</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )}
          />
        )}

        <Button type="submit" disabled={isSubmitting}>
          {book ? t("saveChanges") : t("createBook")}
        </Button>
      </form>
    </Form>
  );
}
