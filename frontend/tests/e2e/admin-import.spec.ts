import path from "path";
import { test, expect } from "@playwright/test";
import { loginAsAdmin } from "./fixtures";

// End-to-end coverage of the highest-risk feature in the product spec: bulk
// Excel/CSV import (upload -> validation preview -> commit -> live progress).
test.describe("Admin: Excel import", () => {
  test("uploads a file, previews it, and completes an import", async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto("/en/admin/import");

    const samplePath = path.resolve(__dirname, "../../../scripts/sample_import.xlsx");
    // Click the real button (rather than reaching straight for the hidden
    // input) so Playwright's actionability wait covers React hydration —
    // driving the hidden input directly can race ahead of the change
    // handler being attached right after navigation.
    const [fileChooser] = await Promise.all([
      page.waitForEvent("filechooser"),
      page.getByRole("button", { name: "Choose file" }).click(),
    ]);
    await fileChooser.setFiles(samplePath);

    await expect(page.getByText("Total rows")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("10", { exact: true }).first()).toBeVisible();

    await page.getByRole("button", { name: "Start import" }).click();

    await expect(page.getByText(/Status: (Completed|Failed)/)).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText("Status: Completed")).toBeVisible();

    // The job must also show up in history immediately.
    await expect(page.getByText("sample_import.xlsx").first()).toBeVisible();
  });
});
