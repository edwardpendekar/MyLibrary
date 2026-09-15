"use client";

import { PUBLIC_BACKEND_ORIGIN } from "./config";
import { ApiError } from "./api-error";
import type { ApiEnvelope } from "@/types/api";

export interface ApiResult<T> {
  data: T;
  meta?: ApiEnvelope<T>["meta"];
}

const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);

function readCookie(name: string): string | null {
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : null;
}

/**
 * Browser-side API client. Calls the Go backend directly (not proxied through
 * a Next.js route) — CORS on the backend allows the frontend origin with
 * credentials, and the auth cookies are SameSite=Strict but still flow because
 * they are scoped to the shared "site" (same registrable domain), just a
 * different port in local dev / a shared origin behind nginx in production.
 *
 * Every non-GET request echoes the (JS-readable) csrf_token cookie back as a
 * header — see backend pkg/csrf and middleware.CSRF for the matching check.
 */
async function request<T>(path: string, init?: RequestInit): Promise<ApiResult<T>> {
  const method = (init?.method ?? "GET").toUpperCase();
  const headers = new Headers(init?.headers);
  if (!SAFE_METHODS.has(method)) {
    const csrfToken = readCookie("csrf_token");
    if (csrfToken) headers.set("X-CSRF-Token", csrfToken);
  }

  const res = await fetch(`${PUBLIC_BACKEND_ORIGIN}${path}`, {
    ...init,
    headers,
    credentials: "include",
  });

  if (res.status === 204) {
    return { data: undefined as T };
  }

  const body = (await res.json()) as ApiEnvelope<T>;
  if (!res.ok || !body.success) {
    throw new ApiError(
      body.error?.message ?? "Request failed",
      body.error?.code ?? "UNKNOWN",
      res.status,
      body.error?.fields
    );
  }
  return { data: body.data, meta: body.meta };
}

function jsonInit(method: string, body?: unknown): RequestInit {
  return {
    method,
    headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  };
}

export const apiClient = {
  get: <T>(path: string) => request<T>(path, { method: "GET" }),
  post: <T>(path: string, body?: unknown) => request<T>(path, jsonInit("POST", body)),
  put: <T>(path: string, body?: unknown) => request<T>(path, jsonInit("PUT", body)),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
  // No Content-Type here on purpose: the browser must set its own
  // multipart/form-data boundary for FormData uploads.
  upload: <T>(path: string, form: FormData, method: "POST" = "POST") =>
    request<T>(path, { method, body: form }),
};
