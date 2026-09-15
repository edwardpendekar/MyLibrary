import { create } from "zustand";
import type { User } from "@/types/api";

interface AuthState {
  user: User | null;
  isLoading: boolean;
  setUser: (user: User | null) => void;
  setLoading: (loading: boolean) => void;
}

// Holds the current user for client components (nav bar, favorite buttons,
// role-gated UI). The source of truth is always the httpOnly cookie on the
// backend; this store is just a client-side cache of `GET /me`, populated by
// the AuthProvider on mount and updated on login/logout.
export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isLoading: true,
  setUser: (user) => set({ user, isLoading: false }),
  setLoading: (isLoading) => set({ isLoading }),
}));
