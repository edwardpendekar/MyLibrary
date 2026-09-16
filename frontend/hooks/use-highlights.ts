"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { highlightsService } from "@/services/highlights.service";
import type { ApiResult } from "@/lib/api-client";

/** Verse IDs the current user has highlighted in this book, synced from the account. */
export function useHighlights(bookId: number, enabled: boolean) {
  return useQuery({
    queryKey: ["highlights", bookId],
    queryFn: () => highlightsService.listByBook(bookId),
    select: (res) => new Set(res.data),
    enabled: enabled && !!bookId,
  });
}

export function useToggleHighlight(bookId: number) {
  const queryClient = useQueryClient();
  const queryKey = ["highlights", bookId];

  return useMutation({
    mutationFn: ({ verseId, highlighted }: { verseId: number; highlighted: boolean }) =>
      highlighted ? highlightsService.remove(verseId) : highlightsService.add(verseId),
    onMutate: async ({ verseId, highlighted }) => {
      await queryClient.cancelQueries({ queryKey });
      // The cache holds the raw query result (ApiResult<number[]>), not the
      // `select`-transformed Set the component reads — mutate that same shape
      // here, or setQueryData silently stores something `select` can't iterate.
      const previous = queryClient.getQueryData<ApiResult<number[]>>(queryKey);
      queryClient.setQueryData<ApiResult<number[]>>(queryKey, (current) => {
        const ids = current?.data ?? [];
        const nextIds = highlighted ? ids.filter((id) => id !== verseId) : [...ids, verseId];
        return { ...current, data: nextIds };
      });
      return { previous };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) queryClient.setQueryData(queryKey, context.previous);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey }),
  });
}
