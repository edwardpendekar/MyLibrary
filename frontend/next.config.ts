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
};

export default createNextIntlPlugin("./i18n/request.ts")(nextConfig);
