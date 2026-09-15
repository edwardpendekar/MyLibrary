import { renderHook, act } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useDebounce } from "./use-debounce";

describe("useDebounce", () => {
  it("returns the initial value immediately", () => {
    const { result } = renderHook(() => useDebounce("first", 300));
    expect(result.current).toBe("first");
  });

  it("only updates after the delay has elapsed", () => {
    vi.useFakeTimers();
    const { result, rerender } = renderHook(({ value }) => useDebounce(value, 300), {
      initialProps: { value: "a" },
    });

    rerender({ value: "b" });
    expect(result.current).toBe("a"); // not yet updated

    act(() => vi.advanceTimersByTime(299));
    expect(result.current).toBe("a"); // still not updated, one ms shy

    act(() => vi.advanceTimersByTime(1));
    expect(result.current).toBe("b"); // now updated

    vi.useRealTimers();
  });

  it("resets the timer on rapid successive changes (only the last value wins)", () => {
    vi.useFakeTimers();
    const { result, rerender } = renderHook(({ value }) => useDebounce(value, 300), {
      initialProps: { value: "a" },
    });

    rerender({ value: "b" });
    act(() => vi.advanceTimersByTime(200));
    rerender({ value: "c" });
    act(() => vi.advanceTimersByTime(200));
    expect(result.current).toBe("a"); // neither b nor c committed yet

    act(() => vi.advanceTimersByTime(100));
    expect(result.current).toBe("c"); // only the final value ever commits

    vi.useRealTimers();
  });
});
