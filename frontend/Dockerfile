# syntax=docker/dockerfile:1
FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
# Our own postinstall hook (copy-pdf-worker.js) needs this present.
COPY scripts ./scripts
# `npm ci` refuses to run here: the lockfile was generated on Windows, and
# several deps (Tailwind's oxide engine, etc.) ship platform-specific optional
# native/WASM binaries that npm only pins for the *generating* host's
# platform — so a strict "must match exactly" install fails on Alpine/Linux
# even though nothing is actually broken. `npm install` still uses the
# lockfile's pinned versions as its source of truth, it just tolerates that
# platform-optional gap instead of rejecting the whole install.
RUN npm install

FROM node:22-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED=1
# Baked into the client bundle at build time — must match the origin browsers
# will actually reach the backend at (typically the nginx-fronted public URL).
ARG NEXT_PUBLIC_BACKEND_ORIGIN=http://localhost:8080
ENV NEXT_PUBLIC_BACKEND_ORIGIN=$NEXT_PUBLIC_BACKEND_ORIGIN
# Origin that book cover/PDF URLs resolve to (see next.config.ts) — usually
# the same public origin as above even when NEXT_PUBLIC_BACKEND_ORIGIN is "".
ARG NEXT_PUBLIC_ASSET_ORIGIN=http://localhost:8080
ENV NEXT_PUBLIC_ASSET_ORIGIN=$NEXT_PUBLIC_ASSET_ORIGIN

RUN npm run build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /app/public ./public
COPY --from=builder --chown=app:app /app/.next/standalone ./
COPY --from=builder --chown=app:app /app/.next/static ./.next/static

USER app
EXPOSE 3000
ENV PORT=3000 HOSTNAME=0.0.0.0
CMD ["node", "server.js"]
