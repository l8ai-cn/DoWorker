FROM mirror.gcr.io/library/node:20-bookworm AS build
WORKDIR /src
RUN corepack enable && corepack prepare pnpm@9 --activate
COPY clients/marketplace-web/package.json clients/marketplace-web/pnpm-lock.yaml ./
RUN pnpm config set registry https://registry.npmmirror.com \
  && pnpm install --frozen-lockfile
ENV NEXT_TELEMETRY_DISABLED=1
COPY clients/marketplace-web/ ./
RUN pnpm exec next build --webpack

FROM gcr.io/distroless/nodejs20-debian12:nonroot
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
COPY --from=build --chown=65532:65532 /src/.next/standalone ./
COPY --from=build --chown=65532:65532 /src/.next/static ./.next/static
EXPOSE 3000
USER nonroot
CMD ["server.js"]
