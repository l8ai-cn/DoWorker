import path from "node:path";
import { fileURLToPath } from "node:url";

const here = path.dirname(fileURLToPath(import.meta.url));
const monorepoRoot = path.resolve(here, "../..");

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",

  // Same pnpm/monorepo NFT fix as clients/web — see that file's comment.
  outputFileTracingRoot: monorepoRoot,

  // `@do-worker/proto` ships raw .ts files (the generated Connect-RPC
  // message classes). Webpack needs to run SWC over them instead of
  // expecting pre-compiled JS. Same reason clients/web lists this in
  // transpilePackages.
  transpilePackages: ["@do-worker/proto"],

  env: {
    NEXT_PUBLIC_PRIMARY_DOMAIN:
      process.env.PRIMARY_DOMAIN || "__PRIMARY_DOMAIN__",
    NEXT_PUBLIC_USE_HTTPS: process.env.USE_HTTPS || "__USE_HTTPS__",
  },

  async rewrites() {
    const primaryDomain = process.env.PRIMARY_DOMAIN;
    const useHttps = process.env.USE_HTTPS === "true";
    const protocol = useHttps ? "https" : "http";
    const backendUrl = primaryDomain
      ? `${protocol}://${primaryDomain}`
      : "http://localhost:10000";
    return [
      { source: "/api/:path*", destination: `${backendUrl}/api/:path*` },
      {
        source: "/:svc/:method",
        has: [{ type: "header", key: "connect-protocol-version" }],
        destination: `${backendUrl}/:svc/:method`,
      },
      { source: "/health", destination: `${backendUrl}/health` },
    ];
  },

  images: {
    remotePatterns: [
      { protocol: "https", hostname: "**" },
      { protocol: "http", hostname: "**" },
    ],
  },
};

export default nextConfig;
