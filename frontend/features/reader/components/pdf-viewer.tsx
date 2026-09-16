"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Bookmark,
  BookmarkCheck,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronUp,
  Maximize,
  Moon,
  Search as SearchIcon,
  Sun,
  ZoomIn,
  ZoomOut,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  useBookmarks,
  useCreateBookmark,
  useDeleteBookmark,
  useLastPosition,
  useSaveLastPosition,
} from "@/hooks/use-bookmarks";
import { usePdfDocument } from "@/hooks/use-pdf-document";
import { useAuthStore } from "@/store/auth-store";
import { useDebounce } from "@/hooks/use-debounce";
import { applyTextHighlight } from "@/lib/pdf-text-highlight";
import type { TextLayer as PdfTextLayer } from "pdfjs-dist";
import styles from "./pdf-viewer.module.css";

const MIN_SCALE = 0.5;
const MAX_SCALE = 3;

export function PdfViewer({ bookId, url }: { bookId: number; url: string }) {
  const t = useTranslations("pdf");
  const { pdf, numPages, isLoading, error } = usePdfDocument(url);
  const user = useAuthStore((s) => s.user);
  const { data: lastPosition } = useLastPosition(bookId);
  const saveLastPosition = useSaveLastPosition();
  const { data: bookmarks } = useBookmarks(bookId);
  const createBookmark = useCreateBookmark(bookId);
  const deleteBookmark = useDeleteBookmark(bookId);

  const [page, setPage] = useState(1);
  const [scale, setScale] = useState(1.2);
  const [isDark, setIsDark] = useState(false);
  const [query, setQuery] = useState("");
  const [matches, setMatches] = useState<number[]>([]);
  const [activeMatchOnPage, setActiveMatchOnPage] = useState(0);
  const [pageMatchCount, setPageMatchCount] = useState(0);
  const debouncedQuery = useDebounce(query, 400);

  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const textLayerRef = useRef<HTMLDivElement | null>(null);
  const textLayerInstanceRef = useRef<PdfTextLayer | null>(null);
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

  // A "current match" index only means something relative to whatever page
  // is showing right now, so both a new search and a page change restart at
  // the first match on that page.
  useEffect(() => {
    // Resetting in response to two independent upstream state changes, not
    // state derivable during render — same intentional pattern as the
    // setMatches([]) reset above.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setActiveMatchOnPage(0);
  }, [debouncedQuery, page]);

  const highlightCurrentLayer = useCallback(() => {
    const textLayer = textLayerInstanceRef.current;
    if (!textLayer) return;
    const { totalMatches, activeElement } = applyTextHighlight(
      textLayer.textDivs,
      textLayer.textContentItemsStr,
      debouncedQuery,
      activeMatchOnPage
    );
    setPageMatchCount(totalMatches);
    activeElement?.scrollIntoView({ block: "center", inline: "center", behavior: "smooth" });
  }, [debouncedQuery, activeMatchOnPage]);

  // Rebuilds the text layer (the invisible, precisely-positioned text runs
  // pdf.js overlays on the canvas) whenever the page or zoom changes, then
  // re-applies whatever search highlight is active.
  useEffect(() => {
    if (!pdf || !textLayerRef.current) return;
    let cancelled = false;
    const container = textLayerRef.current;

    (async () => {
      const pdfjsLib = await import("pdfjs-dist");
      const pdfPage = await pdf.getPage(page);
      const viewport = pdfPage.getViewport({ scale });
      const textContent = await pdfPage.getTextContent();
      if (cancelled) return;

      textLayerInstanceRef.current?.cancel();
      container.replaceChildren();
      // TextLayer's internal font-size calc() expressions read this variable
      // from the container itself (see pdf-viewer.module.css) — the official
      // viewer sets it on an ancestor `.page` element we don't have here.
      container.style.setProperty("--scale-factor", String(scale));
      container.style.setProperty("--total-scale-factor", String(scale));

      const textLayer = new pdfjsLib.TextLayer({ textContentSource: textContent, container, viewport });
      textLayerInstanceRef.current = textLayer;
      try {
        await textLayer.render();
      } catch {
        // cancelled render from a fast page/scale change; safe to ignore
      }
      if (cancelled) return;
      highlightCurrentLayer();
    })();

    return () => {
      cancelled = true;
    };
    // highlightCurrentLayer intentionally excluded: it would rebuild the
    // (expensive) text layer on every keystroke. The separate effect below
    // re-applies highlights to the already-built layer instead.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pdf, page, scale]);

  useEffect(() => {
    highlightCurrentLayer();
  }, [highlightCurrentLayer]);

  const pdfBookmarks = (bookmarks ?? [])
    .filter((b) => b.pdf_page != null)
    .sort((a, b) => (a.pdf_page ?? 0) - (b.pdf_page ?? 0));
  const currentPageBookmark = pdfBookmarks.find((b) => b.pdf_page === page);

  function toggleBookmark() {
    if (currentPageBookmark) {
      deleteBookmark.mutate(currentPageBookmark.id, {
        onSuccess: () => toast.success(t("bookmarkPage")),
      });
    } else {
      createBookmark.mutate(
        { book_id: bookId, pdf_page: page, label: `${t("page")} ${page}` },
        { onSuccess: () => toast.success(t("bookmarkPage")) }
      );
    }
  }

  // Moves to the next/previous occurrence of the search term. Within a page
  // this just steps the active index; crossing a page boundary re-renders
  // that page's text layer, which lands on its first match — landing on a
  // target page's *last* match when going backward would need its text
  // content fetched just to count occurrences before navigating there, which
  // isn't worth the extra round trip for this.
  function goToMatch(direction: 1 | -1) {
    if (matches.length === 0) return;
    if (direction === 1 && activeMatchOnPage + 1 < pageMatchCount) {
      setActiveMatchOnPage((i) => i + 1);
      return;
    }
    if (direction === -1 && activeMatchOnPage > 0) {
      setActiveMatchOnPage((i) => i - 1);
      return;
    }
    const currentIdx = matches.indexOf(page);
    const nextIdx =
      currentIdx === -1
        ? direction === 1
          ? 0
          : matches.length - 1
        : (currentIdx + direction + matches.length) % matches.length;
    setPage(matches[nextIdx]);
  }

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

        {user && (
          <>
            <Button
              variant="outline"
              size="icon-sm"
              aria-label={t("bookmarkPage")}
              aria-pressed={!!currentPageBookmark}
              onClick={toggleBookmark}
            >
              {currentPageBookmark ? (
                <BookmarkCheck className="size-4 text-primary" />
              ) : (
                <Bookmark className="size-4" />
              )}
            </Button>

            {pdfBookmarks.length > 0 && (
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={
                    <Button variant="outline" size="sm">
                      {pdfBookmarks.length}
                    </Button>
                  }
                />
                <DropdownMenuContent align="start">
                  {pdfBookmarks.map((b) => (
                    <DropdownMenuItem key={b.id} onClick={() => setPage(b.pdf_page!)}>
                      {b.label || `${t("page")} ${b.pdf_page}`}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </>
        )}

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

        {matches.length > 0 && (
          <div className="flex items-center gap-0.5">
            <span className="w-14 text-center text-xs tabular-nums text-muted-foreground">
              {pageMatchCount > 0 ? `${activeMatchOnPage + 1}/${pageMatchCount}` : "0/0"}
            </span>
            <Button variant="outline" size="icon-sm" aria-label={t("previousMatch")} onClick={() => goToMatch(-1)}>
              <ChevronUp className="size-4" />
            </Button>
            <Button variant="outline" size="icon-sm" aria-label={t("nextMatch")} onClick={() => goToMatch(1)}>
              <ChevronDown className="size-4" />
            </Button>
          </div>
        )}
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
        <div className={`mx-auto w-fit shadow-lg ${styles.pageWrapper}`}>
          {isLoading && <div className="flex h-96 w-72 items-center justify-center text-muted-foreground">…</div>}
          <canvas ref={canvasRef} className={isDark ? "invert hue-rotate-180" : undefined} />
          <div ref={textLayerRef} className={styles.textLayer} />
        </div>
      </div>
    </div>
  );
}
