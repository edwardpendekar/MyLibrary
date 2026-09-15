import "@testing-library/jest-dom/vitest";

// Node 22+'s experimental native `localStorage` global can shadow jsdom's
// working implementation in some environments, leaving zustand's `persist`
// middleware bound to a stub with no-op/missing methods. Force a real
// in-memory implementation for tests regardless of which one "wins".
class MemoryStorage implements Storage {
  private store = new Map<string, string>();
  get length() {
    return this.store.size;
  }
  clear() {
    this.store.clear();
  }
  getItem(key: string) {
    return this.store.has(key) ? this.store.get(key)! : null;
  }
  key(index: number) {
    return Array.from(this.store.keys())[index] ?? null;
  }
  removeItem(key: string) {
    this.store.delete(key);
  }
  setItem(key: string, value: string) {
    this.store.set(key, String(value));
  }
}

Object.defineProperty(globalThis, "localStorage", {
  value: new MemoryStorage(),
  writable: true,
  configurable: true,
});
