"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import {
  ChevronLeft,
  ChevronRight,
  Maximize,
  Moon,
  Search as SearchIcon,
  Sun,
  ZoomIn,
  ZoomOut,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { usePdfDocument } from "@/hooks/use-pdf-document";
import { useLastPosition, useSaveLastPosition } from "@/hooks/use-bookmarks";
import { useAuthStore } from "@/store/auth-store";
import { useDebounce } from "@/hooks/use-debounce";

const MIN_SCALE = 0.5;
const MAX_SCALE = 3;

export function PdfViewer({ bookId, url }: { bookId: number; url: string }) {
  const t = useTranslations("pdf");
  const { pdf, numPages, isLoading, error } = usePdfDocument(url);
  const user = useAuthStore((s) => s.user);
  const { data: lastPosition } = useLastPosition(bookId);
  const saveLastPosition = useSaveLastPosition();

  const [page, setPage] = useState(1);
  const [scale, setScale] = useState(1.2);
  const [isDark, setIsDark] = useState(false);
  const [query, setQuery] = useState("");
  const [matches, setMatches] = useState<number[]>([]);
  const debouncedQuery = useDebounce(query, 400);

  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const renderTaskRef = useRef<{ cancel: () => void } | null>(null);
  const restoredRef = useRef(false);

  // Resume at the last-read page once, when both the document and the saved
  // position have arrived.
  useEffect(() => {
    if (!restoredRef.current && pdf && lastPosition?.pdf_page) {
      setPage(Math.min(lastPosition.pdf_page, pdf.numPages));
      restoredRef.current = true;
    }
  }, [pdf, lastPosition]);

  const renderPage = useCallback(async () => {
    if (!pdf || !canvasRef.current) return;
    renderTaskRef.current?.cancel();

    const pdfPage = await pdf.getPage(page);
    const viewport = pdfPage.getViewport({ scale });
    const canvas = canvasRef.current;
    const context = canvas.getContext("2d");
    if (!context) return;

    canvas.width = viewport.width;
    canvas.height = viewport.height;

    const task = pdfPage.render({ canvas, canvasContext: context, viewport });
    renderTaskRef.current = task;
    try {
      await task.promise;
    } catch {
      // cancelled render from a fast page/scale change; safe to ignore
    }
  }, [pdf, page, scale]);

  useEffect(() => {
    renderPage();
  }, [renderPage]);

  useEffect(() => {
    if (user && pdf) {
      saveLastPosition.mutate({ book_id: bookId, pdf_page: page });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, bookId, user, !!pdf]);

  useEffect(() => {
    if (!pdf || debouncedQuery.trim().length < 2) {
      // Clearing stale results when the query is cleared/too short, not a
      // derived-state anti-pattern — intentional and safe to disable here.
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setMatches([]);
      return;
    }
    let cancelled = false;
    (async () => {
      const found: number[] = [];
      for (let i = 1; i <= pdf.numPages; i++) {
        if (cancelled) return;
        const pdfPage = await pdf.getPage(i);
        const content = await pdfPage.getTextContent();
        const text = content.items.map((item) => ("str" in item ? item.str : "")).join(" ");
        if (text.toLowerCase().includes(debouncedQuery.toLowerCase())) found.push(i);
      }
      if (!cancelled) setMatches(found);
    })();
    return () => {
      cancelled = true;
    };
  }, [pdf, debouncedQuery]);

  function toggleFullscreen() {
    if (!document.fullscreenElement) {
      containerRef.current?.requestFullscreen();
    } else {
      document.exitFullscreen();
    }
  }

  if (error) {
    return <p className="p-6 text-sm text-destructive">{error}</p>;
  }

  return (
    <div ref={containerRef} className="flex h-[calc(100vh-3.5rem)] flex-col bg-muted">
      <div className="flex flex-wrap items-center gap-2 border-b bg-background px-3 py-2">
        <Button variant="outline" size="icon-sm" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1}>
          <ChevronLeft className="size-4" />
        </Button>
        <span className="text-sm tabular-nums">
          {t("page")} {page} {t("of")} {numPages || "…"}
        </span>
        <Button
          variant="outline"
          size="icon-sm"
          onClick={() => setPage((p) => Math.min(numPages, p + 1))}
          disabled={page >= numPages}
        >
          <ChevronRight className="size-4" />
        </Button>

        <div className="mx-1 h-5 w-px bg-border" />

        <Button variant="outline" size="icon-sm" aria-label={t("zoomOut")} onClick={() => setScale((s) => Math.max(MIN_SCALE, s - 0.2))}>
          <ZoomOut className="size-4" />
        </Button>
        <span className="w-10 text-center text-xs tabular-nums">{Math.round(scale * 100)}%</span>
        <Button variant="outline" size="icon-sm" aria-label={t("zoomIn")} onClick={() => setScale((s) => Math.min(MAX_SCALE, s + 0.2))}>
          <ZoomIn className="size-4" />
        </Button>

        <div className="mx-1 h-5 w-px bg-border" />

        <Button variant="outline" size="icon-sm" aria-label={t("darkMode")} onClick={() => setIsDark((d) => !d)}>
          {isDark ? <Sun className="size-4" /> : <Moon className="size-4" />}
        </Button>
        <Button variant="outline" size="icon-sm" aria-label={t("fullscreen")} onClick={toggleFullscreen}>
          <Maximize className="size-4" />
        </Button>

        <div className="relative ml-auto w-48">
          <SearchIcon className="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t("search")}
            className="h-8 pl-7 text-sm"
          />
        </div>
      </div>

      {matches.length > 0 && (
        <div className="flex items-center gap-1 overflow-x-auto border-b bg-background px-3 py-1.5">
          {matches.map((m) => (
            <Button key={m} variant={m === page ? "secondary" : "ghost"} size="xs" onClick={() => setPage(m)}>
              {m}
            </Button>
          ))}
        </div>
      )}

      <div className="flex-1 overflow-auto p-4">
        <div className="mx-auto w-fit shadow-lg">
          {isLoading && <div className="flex h-96 w-72 items-center justify-center text-muted-foreground">…</div>}
          <canvas ref={canvasRef} className={isDark ? "invert hue-rotate-180" : undefined} />
        </div>
      </div>
    </div>
  );
}
