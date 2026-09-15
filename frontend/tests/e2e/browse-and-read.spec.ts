import { test, expect } from "@playwright/test";
import { ensureGenesisPublished } from "./fixtures";

const BACKEND_ORIGIN = process.env.E2E_BACKEND_ORIGIN ?? "http://localhost:8080";

test.describe("Browse and read", () => {
  test.beforeAll(async ({ request }) => {
    await ensureGenesisPublished(request, BACKEND_ORIGIN);
  });

  test("home page lists a published book", async ({ page }) => {
    await page.goto("/en");
    await expect(page.getByText("Genesis").first()).toBeVisible();
  });

  test("clicking through to a book detail page shows its metadata", async ({ page }) => {
    await page.goto("/en");
    await page.getByText("Genesis").first().click();

    await expect(page).toHaveURL(/\/books\/genesis/);
    await expect(page.getByRole("heading", { name: "Genesis" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Read", exact: true })).toBeVisible();
  });

  test("the verse reader shows chapter text and section headings", async ({ page }) => {
    await page.goto("/en/books/genesis/read/1");

    await expect(page.getByText("In the beginning")).toBeVisible();
    await expect(page.getByText("Creation")).toBeVisible();
    // Bilingual side-by-side text is on by default.
    await expect(page.getByText("Pada mulanya")).toBeVisible();
  });

  test("chapter navigation switches content", async ({ page }) => {
    await page.goto("/en/books/genesis/read/1");
    await page.getByRole("link", { name: "2", exact: true }).click();

    await expect(page).toHaveURL(/\/read\/2/);
    await expect(page.getByText("Thus the heavens and the earth were finished")).toBeVisible();
  });
});
