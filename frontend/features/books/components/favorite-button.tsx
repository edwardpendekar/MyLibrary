"use client";

import { useState } from "react";
import { Heart } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/store/auth-store";
import { useToggleFavorite } from "@/hooks/use-favorites";
import { cn } from "@/lib/utils";

// Tracks its own optimistic boolean instead of reading from a query cache:
// this button is fed server-rendered initial state (from the book detail
// Server Component), which no client query owns, so useToggleFavorite's
// cache-write alone would not update this component's display.
export function FavoriteButton({ bookId, isFavorite: initialIsFavorite }: { bookId: number; isFavorite: boolean }) {
  const t = useTranslations("book");
  const user = useAuthStore((s) => s.user);
  const [isFavorite, setIsFavorite] = useState(initialIsFavorite);
  const toggle = useToggleFavorite(bookId, isFavorite);

  if (!user) return null;

  return (
    <Button
      variant="outline"
      disabled={toggle.isPending}
      onClick={() => {
        setIsFavorite((prev) => !prev);
        toggle.mutate();
      }}
    >
      <Heart className={cn("size-4", isFavorite && "fill-current text-destructive")} />
      {isFavorite ? t("unfavorite") : t("favorite")}
    </Button>
  );
}
