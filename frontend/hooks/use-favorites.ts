"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { favoritesService } from "@/services/favorites.service";
import type { Book } from "@/types/api";

export function useFavorites() {
  return useQuery({
    queryKey: ["favorites"],
    queryFn: () => favoritesService.list(),
    select: (res) => res.data,
  });
}

/** Optimistically flips is_favorite on every cached Book with this id. */
export function useToggleFavorite(bookId: number, isFavorite: boolean) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => (isFavorite ? favoritesService.remove(bookId) : favoritesService.add(bookId)),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: ["books"] });
      const previous = queryClient.getQueriesData<{ data: Book | Book[] }>({ queryKey: ["books"] });

      queryClient.setQueriesData<{ data: Book | Book[] }>({ queryKey: ["books"] }, (old) => {
        if (!old) return old;
        const flip = (b: Book): Book => (b.id === bookId ? { ...b, is_favorite: !isFavorite } : b);
        return { ...old, data: Array.isArray(old.data) ? old.data.map(flip) : flip(old.data) };
      });

      return { previous };
    },
    onError: (_err, _vars, context) => {
      context?.previous.forEach(([key, value]) => queryClient.setQueryData(key, value));
      toast.error("Failed to update favorite");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["books"] });
      queryClient.invalidateQueries({ queryKey: ["favorites"] });
    },
  });
}
