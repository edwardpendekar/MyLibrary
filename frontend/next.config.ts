import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

// Deliberately separate from NEXT_PUBLIC_BACKEND_ORIGIN: that one is the API
// call origin (empty string in the docker-compose/nginx same-origin setup),
// while this is the origin book cover/PDF URLs actually resolve to — the Go
// backend always returns absolute URLs for uploads (see storage.LocalDisk.URL),
// so next/image needs a concrete host to allowlist even when API calls
// themselves are same-origin relative paths.
const assetOrigin = process.env.NEXT_PUBLIC_ASSET_ORIGIN || "http://localhost:8080";
const assetUrl = new URL(assetOrigin);

// The API call origin for connect-src. Empty string (same-origin/nginx setup)
// needs no extra entry since 'self' already covers it.
const backendOrigin = process.env.NEXT_PUBLIC_BACKEND_ORIGIN || "";

const connectSrc = ["'self'", assetUrl.origin, backendOrigin].filter(Boolean).join(" ");
const imgSrc = ["'self'", "data:", "blob:", assetUrl.origin].filter(Boolean).join(" ");

// No nonce-based script-src here: Next.js's own hydration bootstrap script is
// inline, and Base UI (dropdowns/popovers/select) sets inline positioning
// styles at runtime, so 'unsafe-inline' is required for script/style without
// wiring a per-request nonce through middleware. Everything else is locked
// down to 'self' — no third-party scripts, fonts, or frames are used.
const csp = [
  "default-src 'self'",
  `img-src ${imgSrc}`,
  "font-src 'self' data:",
  "script-src 'self' 'unsafe-inline'",
  "style-src 'self' 'unsafe-inline'",
  `connect-src ${connectSrc}`,
  "worker-src 'self' blob:", // pdf.js worker
  "frame-ancestors 'none'",
  "base-uri 'self'",
  "form-action 'self'",
].join("; ");

const nextConfig: NextConfig = {
  output: "standalone",
  images: {
    remotePatterns: [
      {
        protocol: assetUrl.protocol.replace(":", "") as "http" | "https",
        hostname: assetUrl.hostname,
        port: assetUrl.port || undefined,
        pathname: "/uploads/**",
      },
    ],
  },
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [{ key: "Content-Security-Policy", value: csp }],
      },
    ];
  },
};

export default createNextIntlPlugin("./i18n/request.ts")(nextConfig);
