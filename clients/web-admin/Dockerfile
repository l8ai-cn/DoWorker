# syntax=docker/dockerfile:1
FROM node:20-bookworm AS build
WORKDIR /src
RUN corepack enable && corepack prepare pnpm@9 --activate
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages/ packages/
COPY proto/ proto/
COPY clients/web-admin/ clients/web-admin/
# web-admin may import shared packages; install full workspace.
RUN pnpm install --frozen-lockfile
ENV BAZEL_BUILD=standalone
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /src/clients/web-admin
RUN mkdir -p public \
  && node ../../node_modules/next/dist/bin/next build --webpack

FROM gcr.io/distroless/nodejs20-debian12:nonroot
WORKDIR /app
ENV APP_DIR=/app/clients/web-admin
ENV PORT=3001
ENV NODE_ENV=production
COPY --from=build --chown=65532:65532 /src/clients/web-admin/.next/standalone /app/
COPY --from=build --chown=65532:65532 /src/clients/web-admin/.next/static /app/clients/web-admin/.next/static
COPY --from=build --chown=65532:65532 /src/clients/web-admin/public /app/clients/web-admin/public
COPY --from=build --chown=65532:65532 /src/clients/web-admin/entrypoint.mjs /app/entrypoint.mjs
COPY --from=build --chown=65532:65532 /src/clients/web-admin/healthcheck.mjs /app/healthcheck.mjs
EXPOSE 3001
USER nonroot
ENTRYPOINT ["/nodejs/bin/node", "/app/entrypoint.mjs"]
