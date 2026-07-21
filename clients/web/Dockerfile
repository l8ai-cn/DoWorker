FROM docker.io/library/rust:1.88-bookworm AS wasm
WORKDIR /src
RUN mkdir -p /tmp/wasm-pack /tmp/wasm-bindgen \
  && curl --fail --location --silent --show-error --retry 4 --retry-all-errors \
      --connect-timeout 15 \
      https://github.com/rustwasm/wasm-pack/releases/download/v0.13.1/wasm-pack-v0.13.1-x86_64-unknown-linux-musl.tar.gz \
      --output /tmp/wasm-pack.tar.gz \
  && tar --strip-components=1 -xzf /tmp/wasm-pack.tar.gz -C /tmp/wasm-pack \
  && install /tmp/wasm-pack/wasm-pack /usr/local/cargo/bin/wasm-pack \
  && curl --fail --location --silent --show-error --retry 4 --retry-all-errors \
      --connect-timeout 15 \
      https://github.com/wasm-bindgen/wasm-bindgen/releases/download/0.2.105/wasm-bindgen-0.2.105-x86_64-unknown-linux-musl.tar.gz \
      --output /tmp/wasm-bindgen.tar.gz \
  && tar --strip-components=1 -xzf /tmp/wasm-bindgen.tar.gz -C /tmp/wasm-bindgen \
  && install /tmp/wasm-bindgen/wasm-bindgen /tmp/wasm-bindgen/wasm-bindgen-test-runner \
      /tmp/wasm-bindgen/wasm2es6js /usr/local/cargo/bin/ \
  && rm -rf /tmp/wasm-pack /tmp/wasm-pack.tar.gz \
      /tmp/wasm-bindgen /tmp/wasm-bindgen.tar.gz \
  && rustup target add wasm32-unknown-unknown
RUN apt-get update && apt-get install -y --no-install-recommends \
      protobuf-compiler libprotobuf-dev python3 \
  && rm -rf /var/lib/apt/lists/*
COPY proto/ proto/
COPY clients/core/ clients/core/
COPY packages/agent-cloud-wasm/ packages/agent-cloud-wasm/
COPY scripts/build-wasm.sh scripts/seed-rust-proto-stubs.sh scripts/
RUN bash scripts/build-wasm.sh

FROM docker.io/library/node:20-bookworm AS build
WORKDIR /src
ARG NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
ENV npm_config_registry=$NPM_CONFIG_REGISTRY \
    COREPACK_NPM_REGISTRY=$NPM_CONFIG_REGISTRY
RUN corepack enable && corepack prepare pnpm@9 --activate
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages/ packages/
COPY proto/ proto/
COPY clients/web/ clients/web/
COPY --from=wasm /src/packages/agent-cloud-wasm packages/agent-cloud-wasm
RUN pnpm config set registry https://registry.npmmirror.com \
  && pnpm install --frozen-lockfile
ENV STANDALONE=1
ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_BUILD_CPUS=1
ENV NODE_OPTIONS=--max-old-space-size=2048
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
