import type { NextConfig } from "next";
import path from "path";
import { fileURLToPath } from "url";

const here = path.dirname(fileURLToPath(import.meta.url));
const monorepoRoot = path.resolve(here, "../..");

const enableStandalone = process.env.STANDALONE === "1";

const nextConfig: NextConfig = {
  ...(enableStandalone
    ? {
        output: "standalone" as const,
        outputFileTracingRoot: monorepoRoot,
      }
    : {}),

  transpilePackages: ["@agent-cloud/proto"],

  env: {
    NEXT_PUBLIC_PRIMARY_DOMAIN:
      process.env.PRIMARY_DOMAIN || "__PRIMARY_DOMAIN__",
    NEXT_PUBLIC_USE_HTTPS: process.env.USE_HTTPS || "__USE_HTTPS__",
  },

  // Proxy API requests to the backend in development to avoid CORS issues
  async rewrites() {
    const primaryDomain = process.env.PRIMARY_DOMAIN;
    const useHttps = process.env.USE_HTTPS === "true";
    const protocol = useHttps ? "https" : "http";
    const backendUrl = primaryDomain
      ? `${protocol}://${primaryDomain}`
      : "http://localhost:10000";

    return [
      {
        source: "/api/:path*",
        destination: `${backendUrl}/api/:path*`,
      },
      // Connect-RPC: /proto.<svc>.v1.<Service>/<Method> at the root path.
      // path-to-regexp can't match dotted service names in `source`, so gate
      // on the Connect client header (same pattern as clients/web).
      {
        source: "/:svc/:method",
        has: [{ type: "header", key: "connect-protocol-version" }],
        destination: `${backendUrl}/:svc/:method`,
      },
      {
        source: "/health",
        destination: `${backendUrl}/health`,
      },
    ];
  },

  // Allow images from any source during development
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "**",
      },
      {
        protocol: "http",
        hostname: "**",
      },
    ],
  },
};

export default nextConfig;
