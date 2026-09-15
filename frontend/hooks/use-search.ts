"use client";

import { useQuery } from "@tanstack/react-query";
import { searchService } from "@/services/search.service";

export function useSearchVerses(query: string, bookId?: number) {
  return useQuery({
    queryKey: ["search", "verses", query, bookId],
    queryFn: () => searchService.verses(query, { bookId }),
    select: (res) => res.data,
    enabled: query.trim().length > 1,
  });
}

export function useSearchBooks(query: string) {
  return useQuery({
    queryKey: ["search", "books", query],
    queryFn: () => searchService.books(query),
    select: (res) => res.data,
    enabled: query.trim().length > 1,
  });
}
