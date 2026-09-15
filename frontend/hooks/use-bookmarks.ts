"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { bookmarksService, type CreateBookmarkInput, type SaveLastPositionInput } from "@/services/bookmarks.service";

export function useBookmarks(bookId: number) {
  return useQuery({
    queryKey: ["bookmarks", bookId],
    queryFn: () => bookmarksService.listByBook(bookId),
    select: (res) => res.data,
    enabled: !!bookId,
  });
}

export function useCreateBookmark(bookId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateBookmarkInput) => bookmarksService.create(input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["bookmarks", bookId] }),
  });
}

export function useDeleteBookmark(bookId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => bookmarksService.remove(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["bookmarks", bookId] }),
  });
}

/** Debounce this call site-side; it is meant to be cheap but not called per-scroll-frame. */
export function useSaveLastPosition() {
  return useMutation({
    mutationFn: (input: SaveLastPositionInput) => bookmarksService.saveLastPosition(input),
  });
}

export function useLastPosition(bookId: number) {
  return useQuery({
    queryKey: ["last-position", bookId],
    queryFn: () => bookmarksService.lastPosition(bookId),
    select: (res) => res.data,
    enabled: !!bookId,
  });
}
