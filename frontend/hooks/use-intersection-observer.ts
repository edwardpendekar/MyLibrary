import { useEffect, useRef } from "react";

/** Calls onIntersect whenever the returned ref's element enters the viewport. Used to trigger "load more" for infinite scroll. */
export function useIntersectionObserver(onIntersect: () => void, enabled = true) {
  const ref = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!enabled) return;
    const el = ref.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) onIntersect();
      },
      { rootMargin: "400px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled]);

  return ref;
}
