import "server-only";
import { cookies } from "next/headers";

import { BACKEND_ORIGIN } from "./config";
import { ApiError } from "./api-error";
import type { ApiEnvelope } from "@/types/api";

export interface ApiResult<T> {
  data: T;
  meta?: ApiEnvelope<T>["meta"];
}

/**
 * Server Component / Server Action API client. Forwards the visitor's own
 * cookies (access_token) so a server-rendered page can show personalized data
 * (e.g. "is this book favorited?") without a client-side round trip.
 */
async function request<T>(path: string, init?: RequestInit, opts?: { revalidate?: number }): Promise<ApiResult<T>> {
  const cookieStore = await cookies();
  const cookieHeader = cookieStore.toString();

  const res = await fetch(`${BACKEND_ORIGIN}${path}`, {
    ...init,
    headers: {
      ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      ...init?.headers,
    },
    next: opts?.revalidate !== undefined ? { revalidate: opts.revalidate } : undefined,
    cache: opts?.revalidate === undefined ? "no-store" : undefined,
  });

  if (res.status === 204) {
    return { data: undefined as T };
  }

  const body = (await res.json()) as ApiEnvelope<T>;
  if (!res.ok || !body.success) {
    throw new ApiError(body.error?.message ?? "Request failed", body.error?.code ?? "UNKNOWN", res.status, body.error?.fields);
  }
  return { data: body.data, meta: body.meta };
}

export const apiServer = {
  get: <T>(path: string, opts?: { revalidate?: number }) => request<T>(path, { method: "GET" }, opts),
};
