import { defineConfig, devices } from "@playwright/test";

// e2e tests run against a real running stack (frontend + backend + db).
// Point BASE_URL at whichever environment you want to exercise; defaults to
// the local dev frontend.
const baseURL = process.env.E2E_BASE_URL ?? "http://localhost:3000";

export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: false, // tests share seeded admin/book state; keep them ordered
  workers: 1, // same reason — cross-file state (e.g. published books) must not race
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? "github" : "list",
  use: {
    baseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
