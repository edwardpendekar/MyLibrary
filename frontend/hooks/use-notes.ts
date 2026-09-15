"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { notesService } from "@/services/notes.service";

export function useNotes(bookId: number) {
  return useQuery({
    queryKey: ["notes", bookId],
    queryFn: () => notesService.listByBook(bookId),
    select: (res) => res.data,
    enabled: !!bookId,
  });
}

export function useCreateNote(bookId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { verse_id?: number; content: string }) =>
      notesService.create({ book_id: bookId, ...input }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notes", bookId] }),
  });
}

export function useUpdateNote(bookId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, content }: { id: number; content: string }) => notesService.update(id, content),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notes", bookId] }),
  });
}

export function useDeleteNote(bookId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => notesService.remove(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notes", bookId] }),
  });
}
