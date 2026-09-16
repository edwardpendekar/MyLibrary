import { test, expect } from "@playwright/test";
import { ensureGenesisPublished } from "./fixtures";

const BACKEND_ORIGIN = process.env.E2E_BACKEND_ORIGIN ?? "http://localhost:8080";

test.describe("Search", () => {
  test.beforeAll(async ({ request }) => {
    await ensureGenesisPublished(request, BACKEND_ORIGIN);
  });

  test("searching from the home page navigates to results", async ({ page }) => {
    await page.goto("/en");
    await page.getByPlaceholder(/search books, verses/i).fill("beginning");
    await page.keyboard.press("Enter");

    await expect(page).toHaveURL(/\/search\?q=beginning/);
  });

  test("verse search returns a highlighted, relevant result", async ({ page }) => {
    await page.goto("/en/search?q=beginning");

    await expect(page.getByText(/Genesis 1:1/)).toBeVisible({ timeout: 10_000 });
    // Scoped to the Genesis 1:1 result row: the dev dataset has other books
    // whose text also contains "beginning", so an unscoped locator would hit
    // Playwright's strict-mode violation (multiple matches).
    const genesisResult = page.locator("li", { hasText: "Genesis 1:1" });
    // ts_headline wraps the matched term in <b>, rendered as bold text.
    await expect(genesisResult.locator("b", { hasText: "beginning" }).first()).toBeVisible();
  });

  test("book search finds a book by title", async ({ page }) => {
    await page.goto("/en/search?q=Genesis");
    await page.getByRole("tab", { name: "Books" }).click();

    await expect(page.getByText("Genesis").first()).toBeVisible();
  });

  test("an unmatched query shows a no-results message", async ({ page }) => {
    await page.goto("/en/search?q=zzznonexistentqueryzzz");

    await expect(page.getByText(/no results/i)).toBeVisible();
  });
});
