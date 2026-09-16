"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { AlertTriangle } from "lucide-react";

import { Button } from "@/components/ui/button";

// Catches render/data-fetching errors thrown anywhere in this route segment's
// tree. Must be a Client Component (Next.js App Router requirement) — the
// error itself may have come from a Server Component, but the boundary that
// displays it runs client-side so it can offer a "try again" action.
export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const t = useTranslations("common");

  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="mx-auto flex max-w-md flex-col items-center gap-4 px-4 py-24 text-center">
      <AlertTriangle className="size-10 text-destructive" />
      <h1 className="text-xl font-semibold">{t("error")}</h1>
      {error.digest && (
        <p className="text-xs text-muted-foreground">
          Error ID: <code>{error.digest}</code>
        </p>
      )}
      <Button onClick={() => reset()}>{t("retry")}</Button>
    </div>
  );
}
