"use client";

import { useEffect, useRef, useState } from "react";
import type { PDFDocumentLoadingTask, PDFDocumentProxy } from "pdfjs-dist";

export function usePdfDocument(url: string | null) {
  const [pdf, setPdf] = useState<PDFDocumentProxy | null>(null);
  const [numPages, setNumPages] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const taskRef = useRef<PDFDocumentLoadingTask | null>(null);

  useEffect(() => {
    if (!url) return;
    let cancelled = false;

    import("pdfjs-dist").then(async (pdfjsLib) => {
      pdfjsLib.GlobalWorkerOptions.workerSrc = "/pdfjs/pdf.worker.min.mjs";
      const task = pdfjsLib.getDocument({ url });
      taskRef.current = task;
      try {
        const doc = await task.promise;
        if (cancelled) return;
        setPdf(doc);
        setNumPages(doc.numPages);
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load PDF");
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    });

    return () => {
      cancelled = true;
      taskRef.current?.destroy();
      taskRef.current = null;
    };
  }, [url]);

  return { pdf, numPages, isLoading, error };
}
