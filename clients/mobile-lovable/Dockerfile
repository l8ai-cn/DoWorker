# syntax=docker/dockerfile:1
FROM rust:1.88-bookworm AS wasm
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

FROM node:20-bookworm AS build
WORKDIR /src
RUN corepack enable && corepack prepare pnpm@9 --activate
ARG NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
ENV npm_config_registry=$NPM_CONFIG_REGISTRY
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages/ packages/
COPY proto/ proto/
COPY clients/mobile-lovable/ clients/mobile-lovable/
COPY --from=wasm /src/packages/do-worker-wasm packages/do-worker-wasm
RUN pnpm install --frozen-lockfile
ARG VITE_DO_WORKER_API_URL=https://dowork.l8ai.cn
ENV VITE_DO_WORKER_API_URL=$VITE_DO_WORKER_API_URL
WORKDIR /src/clients/mobile-lovable
RUN pnpm build

FROM gcr.io/distroless/nodejs20-debian12:nonroot
WORKDIR /app
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000
COPY --from=build --chown=65532:65532 /src/clients/mobile-lovable/.output /app
EXPOSE 3000
USER nonroot
ENTRYPOINT ["/nodejs/bin/node", "/app/server/index.mjs"]
