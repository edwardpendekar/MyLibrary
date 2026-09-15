import fs from "fs";
import path from "path";
import type { APIRequestContext, Page } from "@playwright/test";

export const ADMIN_EMAIL = "admin@bookreader.local";
export const ADMIN_PASSWORD = "ChangeMe123!";

export async function loginAsAdmin(page: Page) {
  await page.goto("/en/login");
  await page.getByLabel("Email").fill(ADMIN_EMAIL);
  await page.getByLabel("Password").fill(ADMIN_PASSWORD);
  await page.locator("form").getByRole("button", { name: "Log in" }).click();
  await page.waitForURL(/\/en$/);
}

/**
 * Fast, UI-independent setup for specs that need a published book to already
 * exist (browse/read/search) without depending on admin-import.spec.ts having
 * run first. Uses the backend API directly via Playwright's request context,
 * which shares cookies with `request` but not with any `page` in the test —
 * callers that also need a logged-in `page` should log in separately.
 */
export async function ensureGenesisPublished(request: APIRequestContext, backendOrigin: string) {
  const login = await request.post(`${backendOrigin}/api/v1/auth/login`, {
    data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
  });
  if (!login.ok()) throw new Error(`admin login failed: ${login.status()}`);

  // The backend's double-submit CSRF check requires this header on every
  // non-GET request once the session cookie is set — see pkg/csrf.
  const csrfToken = await readCsrfCookie(request, backendOrigin);
  const csrfHeaders = { "X-CSRF-Token": csrfToken };

  const samplePath = path.resolve(__dirname, "../../../scripts/sample_import.xlsx");
  const upload = await request.post(`${backendOrigin}/api/v1/admin/import/upload`, {
    multipart: {
      file: {
        name: "sample_import.xlsx",
        mimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        buffer: fs.readFileSync(samplePath),
      },
    },
    headers: csrfHeaders,
  });
  if (!upload.ok()) throw new Error(`import upload failed: ${upload.status()}`);
  const preview = await upload.json();

  const commit = await request.post(
    `${backendOrigin}/api/v1/admin/import/${preview.data.import_log_id}/commit`,
    { data: { mode: "insert" }, headers: csrfHeaders }
  );
  if (!commit.ok()) throw new Error(`import commit failed: ${commit.status()}`);

  // Poll until the import finishes (or is already done from a prior run).
  for (let i = 0; i < 20; i++) {
    const status = await request.get(
      `${backendOrigin}/api/v1/admin/import/${preview.data.import_log_id}`
    );
    const body = await status.json();
    if (["completed", "failed", "rolled_back"].includes(body.data.status)) break;
    await new Promise((r) => setTimeout(r, 500));
  }

  const books = await request.get(`${backendOrigin}/api/v1/admin/books?q=Genesis`);
  const { data } = await books.json();
  const genesis = data.find((b: { slug: string }) => b.slug === "genesis");
  if (!genesis) throw new Error("Genesis was not created by the import");

  if (genesis.status !== "published") {
    await request.put(`${backendOrigin}/api/v1/admin/books/${genesis.id}`, {
      data: { title: genesis.title, status: "published" },
      headers: csrfHeaders,
    });
  }
}

async function readCsrfCookie(request: APIRequestContext, backendOrigin: string): Promise<string> {
  const cookies = await request.storageState();
  const match = cookies.cookies.find(
    (c) => c.name === "csrf_token" && backendOrigin.includes(c.domain.replace(/^\./, ""))
  );
  if (!match) throw new Error("csrf_token cookie not found after login");
  return match.value;
}
