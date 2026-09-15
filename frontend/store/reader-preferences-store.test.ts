import { beforeEach, describe, expect, it } from "vitest";
import { useReaderPreferencesStore } from "./reader-preferences-store";

describe("useReaderPreferencesStore", () => {
  beforeEach(() => {
    useReaderPreferencesStore.setState({ fontSize: 18, translation: "both" });
  });

  it("increases font size by 2px per step", () => {
    useReaderPreferencesStore.getState().increaseFontSize();
    expect(useReaderPreferencesStore.getState().fontSize).toBe(20);
  });

  it("clamps font size at the configured maximum (28px)", () => {
    useReaderPreferencesStore.setState({ fontSize: 28 });
    useReaderPreferencesStore.getState().increaseFontSize();
    expect(useReaderPreferencesStore.getState().fontSize).toBe(28);
  });

  it("clamps font size at the configured minimum (14px)", () => {
    useReaderPreferencesStore.setState({ fontSize: 14 });
    useReaderPreferencesStore.getState().decreaseFontSize();
    expect(useReaderPreferencesStore.getState().fontSize).toBe(14);
  });

  it("updates the translation display mode", () => {
    useReaderPreferencesStore.getState().setTranslation("id");
    expect(useReaderPreferencesStore.getState().translation).toBe("id");
  });
});
