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
    await page.locator('input[type="file"]').setInputFiles(samplePath);

    await expect(page.getByText("Total rows")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("10", { exact: true }).first()).toBeVisible();

    await page.getByRole("button", { name: "Start import" }).click();

    await expect(page.getByText(/Status: (completed|failed)/)).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText("Status: completed")).toBeVisible();

    // The job must also show up in history immediately.
    await expect(page.getByText("sample_import.xlsx").first()).toBeVisible();
  });
});
