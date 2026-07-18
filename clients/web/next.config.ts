import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";
import path from "path";
import { fileURLToPath } from "url";

const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");

// Turbopack's `root` must be the monorepo root (where node_modules/.pnpm lives),
// NOT the project directory. Previously when `web/` was top-level, Next could
// auto-infer this; after moving under `clients/web/`, we must pin it.
const here = path.dirname(fileURLToPath(import.meta.url));
const monorepoRoot = path.resolve(here, "../..");

function getDevProxyTarget(): string {
  if (process.env.API_PROXY_TARGET) return process.env.API_PROXY_TARGET;

  const domain = process.env.PRIMARY_DOMAIN;
  if (domain) {
    const protocol = process.env.USE_HTTPS === "true" ? "https" : "http";
    return `${protocol}://${domain}`;
  }

  return "http://localhost:10000";
}

// `output: 'standalone'` packages the server + transitive node_modules
// into `.next/standalone/` for slim Docker images (see clients/web/Dockerfile).
// Dev (`next dev`) keeps the default output; production scripts set STANDALONE=1.
const enableStandalone = process.env.STANDALONE === "1";

const nextConfig: NextConfig = {
  ...(enableStandalone
    ? {
        output: "standalone" as const,
        // pnpm monorepo: NFT must walk the virtual store at repo root.
        outputFileTracingRoot: monorepoRoot,
      }
    : {}),

  // Type checks live in `pnpm run web:typecheck`. Don't re-run them
  // inside `next build` — the Next.js build path hits a stricter
  // JSX-inference pass that flags pre-existing implicit-any sites the
  // top-level `tsc --noEmit` already accepts.
  typescript: { ignoreBuildErrors: true },

  // Workspace packages ship raw .ts sources; SWC must transpile them.
  transpilePackages: [
    "@do-worker/agent-ui",
    "@do-worker/service-runtime",
    "@do-worker/service-interface",
    "@do-worker/proto",
  ],

  webpack: (config, { isServer }) => {
    config.experiments = {
      ...config.experiments,
      asyncWebAssembly: true,
    };
    config.output.webassemblyModuleFilename = isServer
      ? "./../static/wasm/[modulehash].wasm"
      : "static/wasm/[modulehash].wasm";
    return config;
  },
  allowedDevOrigins: process.env.ALLOWED_DEV_ORIGINS
    ? process.env.ALLOWED_DEV_ORIGINS.split(",")
    : [],

  // Ensure standalone build includes blog markdown files
  outputFileTracingIncludes: {
    "/blog/[slug]": ["./src/content/blog/**/*.md"],
    "/blog": ["./src/content/blog/**/*.md"],
  },

  // Required for next-intl plugin to resolve config in Turbopack dev mode.
  // `root` must point to the monorepo root so Turbopack can find the pnpm
  // virtual store at `<root>/node_modules/.pnpm/`.
  turbopack: {
    root: monorepoRoot,
  },

  // =============================================================================
  // Unified Domain Configuration
  // 将 PRIMARY_DOMAIN / USE_HTTPS 映射为 NEXT_PUBLIC_* 变量
  // 这样配置文件中可以统一使用 PRIMARY_DOMAIN，与 Backend/Relay 保持一致
  // =============================================================================
  env: {
    // 使用占位符，运行时由 entrypoint.mjs 替换为实际值
    // 构建时直接读 process.env 会被 Next.js 内联求值，导致占位符替换失效
    NEXT_PUBLIC_PRIMARY_DOMAIN:
      process.env.PRIMARY_DOMAIN || "__PRIMARY_DOMAIN__",
    NEXT_PUBLIC_USE_HTTPS: process.env.USE_HTTPS || "__USE_HTTPS__",
    NEXT_PUBLIC_POSTHOG_KEY:
      process.env.POSTHOG_KEY || "__POSTHOG_KEY__",
    NEXT_PUBLIC_POSTHOG_HOST:
      process.env.POSTHOG_HOST || "__POSTHOG_HOST__",
    // Build-time gate for test-only UI surfaces (e.g. e2e-echo credential
    // form). Inlined by Next.js DefinePlugin so the `if (process.env.
    // NEXT_PUBLIC_E2E === "true")` branches are dead-code-eliminated in
    // production builds. Set to "true" only in dev/e2e (see
    // deploy/dev/lib/bootstrap.sh) — defaults to empty string in prod,
    // never "true" by accident.
    NEXT_PUBLIC_E2E: process.env.NEXT_PUBLIC_E2E === "true" ? "true" : "",
  },

  async headers() {
    return [
      {
        source: "/:path*.wasm",
        headers: [
          { key: "Content-Type", value: "application/wasm" },
          { key: "Cache-Control", value: "public, max-age=31536000, immutable" },
        ],
      },
    ];
  },

  async rewrites() {
    if (process.env.NODE_ENV === "development") {
      const proxyTarget = getDevProxyTarget();
      const marketplaceTarget =
        process.env.MARKETPLACE_API_PROXY_TARGET || "http://localhost:10022";
      console.log(`[Next.js] API proxy enabled: /api/* + /v1/* + /proto.* + /health → ${proxyTarget}`);
      return [
        {
          source: "/api/marketplace/v1/:path*",
          destination: `${marketplaceTarget}/api/marketplace/v1/:path*`,
        },
        {
          source: "/api/:path*",
          destination: `${proxyTarget}/api/:path*`,
        },
        // Session REST API is mounted at bare `/v1/*` on the backend (e.g.
        // /v1/virtual-keys, /v1/usage/quota-report). quotaApi.ts reaches it
        // via plain fetch without the connect-protocol-version header, so the
        // Connect `/:svc/:method` rewrite below can't catch these — proxy the
        // whole `/v1` prefix explicitly. Traefik already routes `/v1` in prod.
        {
          source: "/v1/:path*",
          destination: `${proxyTarget}/v1/:path*`,
        },
        // Connect-RPC procedures use the path `/proto.<svc>.v1.Service/Method`.
        // Next.js path-to-regexp doesn't tolerate escaped dots in `source`,
        // so match by the `connect-protocol-version` header that every
        // Connect client sends. Browsers without this header (regular page
        // requests) don't match — keeps the marketing routes intact.
        {
          source: "/:svc/:method",
          has: [{ type: "header", key: "connect-protocol-version" }],
          destination: `${proxyTarget}/:svc/:method`,
        },
        {
          source: "/health",
          destination: `${proxyTarget}/health`,
        },
      ];
    }

    return [];
  },
};

export default withNextIntl(nextConfig);
