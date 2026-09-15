"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { adminCategoriesService, adminLanguagesService } from "@/services/admin/reference.service";

export function useAdminLanguages() {
  return useQuery({
    queryKey: ["admin", "languages"],
    queryFn: () => adminLanguagesService.list(),
    select: (res) => res.data,
  });
}

export function useCreateLanguage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: adminLanguagesService.create,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin", "languages"] }),
  });
}

export function useDeleteLanguage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => adminLanguagesService.remove(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin", "languages"] }),
  });
}

export function useAdminCategories() {
  return useQuery({
    queryKey: ["admin", "categories"],
    queryFn: () => adminCategoriesService.list(),
    select: (res) => res.data,
  });
}

export function useCreateCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: adminCategoriesService.create,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin", "categories"] }),
  });
}

export function useDeleteCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => adminCategoriesService.remove(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin", "categories"] }),
  });
}
