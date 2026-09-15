"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { adminBooksService, type UpsertBookInput } from "@/services/admin/books.service";

export function useAdminBooks(q?: string) {
  return useQuery({
    queryKey: ["admin", "books", q],
    queryFn: () => adminBooksService.list({ q }),
    select: (res) => res.data,
  });
}

export function useAdminBook(id: number) {
  return useQuery({
    queryKey: ["admin", "books", id],
    queryFn: () => adminBooksService.get(id),
    select: (res) => res.data,
    enabled: !!id,
  });
}

export function useCreateBook() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpsertBookInput) => adminBooksService.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "books"] });
      toast.success("Book created");
    },
  });
}

export function useUpdateBook(id: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpsertBookInput) => adminBooksService.update(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "books"] });
      toast.success("Book updated");
    },
  });
}

export function useDeleteBook() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => adminBooksService.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "books"] });
      toast.success("Book deleted");
    },
  });
}

export function useUploadCover(id: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => adminBooksService.uploadCover(id, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "books", id] });
      toast.success("Cover uploaded");
    },
  });
}

export function useUploadPdf(id: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => adminBooksService.uploadPdf(id, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "books", id] });
      toast.success("PDF uploaded");
    },
  });
}
