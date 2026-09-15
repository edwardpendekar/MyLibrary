import { create } from "zustand";
import { persist } from "zustand/middleware";

interface HighlightsState {
  // verseId -> highlighted
  highlighted: Record<number, boolean>;
  toggle: (verseId: number) => void;
  isHighlighted: (verseId: number) => boolean;
}

// Highlights are device-local only (no highlights table in the backend schema
// yet) — persisted to localStorage so they survive reloads on the same device.
// A future iteration could sync these to the account via a dedicated endpoint.
export const useHighlightsStore = create<HighlightsState>()(
  persist(
    (set, get) => ({
      highlighted: {},
      toggle: (verseId) =>
        set((s) => ({ highlighted: { ...s.highlighted, [verseId]: !s.highlighted[verseId] } })),
      isHighlighted: (verseId) => !!get().highlighted[verseId],
    }),
    { name: "verse-highlights" }
  )
);
