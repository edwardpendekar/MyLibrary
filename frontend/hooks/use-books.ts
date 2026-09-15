"use client";

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";

import { booksService, type ListBooksParams } from "@/services/books.service";

export function useInfiniteBooks(params: Omit<ListBooksParams, "cursor"> = {}) {
  return useInfiniteQuery({
    queryKey: ["books", params],
    queryFn: ({ pageParam }) => booksService.list({ ...params, cursor: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => (lastPage.meta?.has_more ? lastPage.meta?.next_cursor : undefined),
  });
}

export function usePopularBooks(limit = 10) {
  return useQuery({
    queryKey: ["books", "popular", limit],
    queryFn: () => booksService.popular(limit),
    select: (res) => res.data,
  });
}

export function useBook(slug: string) {
  return useQuery({
    queryKey: ["books", "detail", slug],
    queryFn: () => booksService.bySlug(slug),
    select: (res) => res.data,
    enabled: !!slug,
  });
}

export function useChapters(bookId: number) {
  return useQuery({
    queryKey: ["books", bookId, "chapters"],
    queryFn: () => booksService.chapters(bookId),
    select: (res) => res.data,
    enabled: !!bookId,
  });
}

export function useChapterContent(bookId: number, chapterNumber: number) {
  return useQuery({
    queryKey: ["books", bookId, "chapters", chapterNumber],
    queryFn: () => booksService.chapterContent(bookId, chapterNumber),
    select: (res) => res.data,
    enabled: !!bookId && !!chapterNumber,
  });
}
