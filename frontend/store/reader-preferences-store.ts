import { create } from "zustand";
import { persist } from "zustand/middleware";

export type ReaderTranslation = "en" | "id" | "both";

interface ReaderPreferencesState {
  fontSize: number; // px
  translation: ReaderTranslation;
  increaseFontSize: () => void;
  decreaseFontSize: () => void;
  setTranslation: (t: ReaderTranslation) => void;
}

const MIN_FONT = 14;
const MAX_FONT = 28;

// Purely local reading preferences (font size, EN/ID/both display mode).
// Persisted to localStorage since they are per-device UI state, not account
// data — they intentionally do not sync through the backend.
export const useReaderPreferencesStore = create<ReaderPreferencesState>()(
  persist(
    (set) => ({
      fontSize: 18,
      translation: "both",
      increaseFontSize: () => set((s) => ({ fontSize: Math.min(MAX_FONT, s.fontSize + 2) })),
      decreaseFontSize: () => set((s) => ({ fontSize: Math.max(MIN_FONT, s.fontSize - 2) })),
      setTranslation: (translation) => set({ translation }),
    }),
    { name: "reader-preferences" }
  )
);
