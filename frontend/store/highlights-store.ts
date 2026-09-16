import { create } from "zustand";
import { persist } from "zustand/middleware";

interface HighlightsState {
  // verseId -> highlighted
  highlighted: Record<number, boolean>;
  toggle: (verseId: number) => void;
  isHighlighted: (verseId: number) => boolean;
}

// Fallback for guests only. Logged-in users get account-synced highlights via
// the /verses/:id/highlight endpoints (see hooks/use-highlights.ts); this
// localStorage-backed store exists purely so an unauthenticated reader can
// still highlight verses on their current device.
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
