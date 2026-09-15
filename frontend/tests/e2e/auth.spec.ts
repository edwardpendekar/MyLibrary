import { test, expect } from "@playwright/test";

// Uses the seed admin account from migration 000019_seed_reference_data.
// Rotate this password immediately in any non-local environment (see
// docs/deployment.md) — tests here assume the local/dev default only.
const ADMIN_EMAIL = "admin@bookreader.local";
const ADMIN_PASSWORD = "ChangeMe123!";

test.describe("Authentication", () => {
  test("rejects an invalid login", async ({ page }) => {
    await page.goto("/en/login");
    await page.getByLabel("Email").fill("nobody@example.com");
    await page.getByLabel("Password").fill("wrong-password");
    await page.locator("form").getByRole("button", { name: "Log in" }).click();

    await expect(page).toHaveURL(/\/login/);
  });

  test("logs the seeded admin in and shows the admin nav link", async ({ page }) => {
    await page.goto("/en/login");
    await page.getByLabel("Email").fill(ADMIN_EMAIL);
    await page.getByLabel("Password").fill(ADMIN_PASSWORD);
    await page.locator("form").getByRole("button", { name: "Log in" }).click();

    await expect(page).toHaveURL(/\/en$/);
    await expect(page.getByRole("link", { name: /admin/i })).toBeVisible();
  });

  test("logging out clears the session", async ({ page }) => {
    await page.goto("/en/login");
    await page.getByLabel("Email").fill(ADMIN_EMAIL);
    await page.getByLabel("Password").fill(ADMIN_PASSWORD);
    await page.locator("form").getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL(/\/en$/);

    await page.getByRole("button", { name: "User menu" }).click();
    await page.getByText("Log out").click();

    await expect(page.getByRole("link", { name: "Log in" })).toBeVisible();
  });
});
