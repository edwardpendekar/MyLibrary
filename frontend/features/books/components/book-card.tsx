"use client";

import { useTranslations } from "next-intl";
import { Heart } from "lucide-react";

import { Link } from "@/i18n/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { BookCover } from "@/features/books/components/book-cover";
import { useAuthStore } from "@/store/auth-store";
import { useToggleFavorite } from "@/hooks/use-favorites";
import { cn } from "@/lib/utils";
import type { Book } from "@/types/api";

export function BookCard({ book, priority = false }: { book: Book; priority?: boolean }) {
  const t = useTranslations("book");
  const user = useAuthStore((s) => s.user);
  const toggleFavorite = useToggleFavorite(book.id, !!book.is_favorite);

  return (
    <div className="group relative">
      <Link href={`/books/${book.slug}`} className="block">
        <BookCover src={book.cover_url} alt={book.title} priority={priority} />
        <div className="mt-2 space-y-0.5">
          <p className="line-clamp-1 text-sm font-medium">{book.title}</p>
          {book.author && (
            <p className="line-clamp-1 text-xs text-muted-foreground">{t("by", { author: book.author })}</p>
          )}
          <div className="flex items-center gap-1.5 pt-0.5">
            {book.language && (
              <Badge variant="secondary" className="text-[10px]">
                {book.language.code.toUpperCase()}
              </Badge>
            )}
          </div>
        </div>
      </Link>

      {user && (
        <Button
          variant="secondary"
          size="icon-sm"
          className="absolute right-2 top-2 opacity-0 shadow transition-opacity group-hover:opacity-100 data-[active=true]:opacity-100"
          data-active={book.is_favorite}
          aria-label={book.is_favorite ? t("unfavorite") : t("favorite")}
          onClick={(e) => {
            e.preventDefault();
            toggleFavorite.mutate();
          }}
        >
          <Heart className={cn("size-3.5", book.is_favorite && "fill-current text-destructive")} />
        </Button>
      )}
    </div>
  );
}
