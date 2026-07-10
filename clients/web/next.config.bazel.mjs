import createNextIntlPlugin from "next-intl/plugin";
import path from "node:path";
import { fileURLToPath } from "node:url";

const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");

const here = path.dirname(fileURLToPath(import.meta.url));
const monorepoRoot = path.resolve(here, "../..");

function getDevProxyTarget() {
  if (process.env.API_PROXY_TARGET) return process.env.API_PROXY_TARGET;

  const domain = process.env.PRIMARY_DOMAIN;
  if (domain) {
    const protocol = process.env.USE_HTTPS === "true" ? "https" : "http";
    return `${protocol}://${domain}`;
  }

  return "http://localhost:10000";
}

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",

  // pnpm + monorepo: NFT must walk the virtual store at the repo root,
  // otherwise transitively-imported deps disappear from .next/standalone/.
  // See vercel/next.js#33895, #48017.
  outputFileTracingRoot: monorepoRoot,

  outputFileTracingIncludes: {
    "/blog/[slug]": ["./src/content/blog/**/*.md"],
    "/blog": ["./src/content/blog/**/*.md"],
  },

  typescript: { ignoreBuildErrors: true },

  transpilePackages: [
    "@do-worker/service-runtime",
    "@do-worker/service-interface",
    // Internal npm package mounted by Bazel; ships .ts sources so the
    // .next/standalone build relies on Next's SWC pipeline to transpile.
    "@do-worker/proto",
  ],

  webpack: (config, { isServer }) => {
    config.experiments = { ...config.experiments, asyncWebAssembly: true };
    config.output.webassemblyModuleFilename = isServer
      ? "./../static/wasm/[modulehash].wasm"
      : "static/wasm/[modulehash].wasm";
    return config;
  },

  allowedDevOrigins: process.env.ALLOWED_DEV_ORIGINS
    ? process.env.ALLOWED_DEV_ORIGINS.split(",")
    : [],

  turbopack: { root: monorepoRoot },

  env: {
    NEXT_PUBLIC_PRIMARY_DOMAIN:
      process.env.PRIMARY_DOMAIN || "__PRIMARY_DOMAIN__",
    NEXT_PUBLIC_USE_HTTPS: process.env.USE_HTTPS || "__USE_HTTPS__",
    NEXT_PUBLIC_POSTHOG_KEY: process.env.POSTHOG_KEY || "__POSTHOG_KEY__",
    NEXT_PUBLIC_POSTHOG_HOST: process.env.POSTHOG_HOST || "__POSTHOG_HOST__",
  },

  async rewrites() {
    if (process.env.NODE_ENV === "development") {
      const proxyTarget = getDevProxyTarget();
      console.log(`[Next.js] API proxy enabled: /api/* + /proto.* + /health -> ${proxyTarget}`);
      return [
        { source: "/api/:path*", destination: `${proxyTarget}/api/:path*` },
        {
          source: "/:svc/:method",
          has: [{ type: "header", key: "connect-protocol-version" }],
          destination: `${proxyTarget}/:svc/:method`,
        },
        { source: "/health", destination: `${proxyTarget}/health` },
      ];
    }
    return [];
  },
};

export default withNextIntl(nextConfig);
