FROM docker.io/library/rust:1.88-bookworm AS wasm
WORKDIR /src
RUN curl https://rustwasm.github.io/wasm-pack/installer/init.sh -sSf | sh \
  && cargo install wasm-bindgen-cli --version 0.2.105 --locked \
  && rustup target add wasm32-unknown-unknown
RUN apt-get update && apt-get install -y --no-install-recommends \
      protobuf-compiler libprotobuf-dev python3 \
  && rm -rf /var/lib/apt/lists/*
COPY proto/ proto/
COPY clients/core/ clients/core/
COPY packages/do-worker-wasm/ packages/do-worker-wasm/
COPY scripts/build-wasm.sh scripts/seed-rust-proto-stubs.sh scripts/
RUN bash scripts/build-wasm.sh

FROM docker.io/library/node:20-bookworm AS build
WORKDIR /src
RUN corepack enable && corepack prepare pnpm@9 --activate
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages/ packages/
COPY proto/ proto/
COPY clients/web/ clients/web/
COPY --from=wasm /src/packages/do-worker-wasm packages/do-worker-wasm
RUN pnpm config set registry https://registry.npmmirror.com \
  && pnpm install --frozen-lockfile
ENV BAZEL_BUILD=standalone
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /src/clients/web
RUN node ../../node_modules/next/dist/bin/next build --webpack

FROM gcr.io/distroless/nodejs20-debian12:nonroot
WORKDIR /app
ENV APP_DIR=/app/clients/web
ENV NODE_ENV=production
COPY --from=build --chown=65532:65532 /src/clients/web/.next/standalone /app/
COPY --from=build --chown=65532:65532 /src/clients/web/.next/static /app/clients/web/.next/static
COPY --from=build --chown=65532:65532 /src/clients/web/public /app/clients/web/public
COPY --from=build --chown=65532:65532 /src/clients/web/entrypoint.mjs /app/entrypoint.mjs
COPY --from=build --chown=65532:65532 /src/clients/web/healthcheck.mjs /app/healthcheck.mjs
EXPOSE 3000
USER nonroot
ENTRYPOINT ["/nodejs/bin/node", "/app/entrypoint.mjs"]
