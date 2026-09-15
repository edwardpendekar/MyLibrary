// Server-side calls hit the backend directly by its (possibly internal-Docker)
// origin; the browser needs a publicly reachable origin, which may differ.
export const BACKEND_ORIGIN =
  process.env.BACKEND_INTERNAL_ORIGIN ??
  process.env.NEXT_PUBLIC_BACKEND_ORIGIN ??
  "http://localhost:8080";

export const PUBLIC_BACKEND_ORIGIN =
  process.env.NEXT_PUBLIC_BACKEND_ORIGIN ?? "http://localhost:8080";
