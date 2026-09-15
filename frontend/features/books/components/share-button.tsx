"use client";

import { Share2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";

// path is a site-relative path (e.g. "/books/genesis"); the absolute URL is
// resolved from window.location at click time so this works in any environment
// (localhost, staging, production) without a server-computed origin.
export function ShareButton({ title, path }: { title: string; path: string }) {
  const t = useTranslations("book");

  return (
    <Button
      variant="ghost"
      onClick={async () => {
        const url = `${window.location.origin}${path}`;
        if (navigator.share) {
          await navigator.share({ title, url }).catch(() => {});
          return;
        }
        await navigator.clipboard.writeText(url);
        toast.success("Link copied");
      }}
    >
      <Share2 className="size-4" />
      {t("share")}
    </Button>
  );
}
