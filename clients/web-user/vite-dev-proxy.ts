import { execFileSync } from "node:child_process";
import type { ProxyOptions } from "vite";

let cachedToken: string | null | undefined;

function resolveToken(host: string): string | null {
  if (cachedToken !== undefined) return cachedToken;

  if (process.env.DO_WORKER_AUTH_TOKEN) {
    cachedToken = process.env.DO_WORKER_AUTH_TOKEN;
    return cachedToken;
  }

  try {
    const output = execFileSync(
      "databricks",
      ["auth", "token", "--host", host, "--output", "json"],
      {
        encoding: "utf8",
        stdio: ["ignore", "pipe", "pipe"],
      },
    );
    const tokenResponse = JSON.parse(output) as { access_token?: string };
    cachedToken = tokenResponse.access_token ?? null;
  } catch {
    cachedToken = null;
  }

  return cachedToken;
}

function configureProxy(target: string, useAuth: boolean): NonNullable<ProxyOptions["configure"]> {
  const parsed = new URL(target);
  const host = parsed.origin;
  const basePath = parsed.pathname.replace(/\/$/, "");

  return (proxy) => {
    proxy.on("proxyReq", (proxyReq) => {
      if (basePath) proxyReq.path = `${basePath}${proxyReq.path}`;
      if (useAuth) {
        const token = resolveToken(host);
        if (token) proxyReq.setHeader("Authorization", `Bearer ${token}`);
      }
    });

    proxy.on("proxyReqWs", (proxyReq) => {
      if (basePath) proxyReq.path = `${basePath}${proxyReq.path}`;
      if (useAuth) {
        const token = resolveToken(host);
        if (token) proxyReq.setHeader("Authorization", `Bearer ${token}`);
      }
    });

    proxy.on("proxyRes", (proxyRes, _req, res) => {
      const contentType = proxyRes.headers["content-type"] ?? "";
      if (typeof contentType === "string" && contentType.includes("text/event-stream")) {
        setImmediate(() => res.flushHeaders());
      }
    });
  };
}

function createProxyRoutes(target: string, useAuth: boolean): Record<string, ProxyOptions> {
  const origin = new URL(target).origin;
  const configure = configureProxy(target, useAuth);

  return {
    "/v1": { target: origin, changeOrigin: true, ws: true, configure },
    "/api": { target: origin, changeOrigin: true, configure },
    "/proto": { target: origin, changeOrigin: true, configure },
    "/auth": { target: origin, changeOrigin: true, configure },
    "/health": { target: origin, changeOrigin: true, configure },
  };
}

export function createDevProxyConfig(target: string): Record<string, ProxyOptions> {
  const parsed = new URL(target);
  const useAuth =
    !!process.env.DO_WORKER_AUTH_TOKEN ||
    parsed.hostname.endsWith(".databricks.com") ||
    parsed.hostname.endsWith(".azuredatabricks.net");

  if (!useAuth) {
    console.log(`[dev-proxy] target=${target}`);
    return createProxyRoutes(target, false);
  }

  const token = resolveToken(parsed.origin);
  if (!token) {
    console.error(
      `\n[dev-proxy] ERROR: No auth token for ${parsed.origin}.\n` +
        `  Set DO_WORKER_AUTH_TOKEN or run:  databricks auth login --host ${parsed.origin}\n`,
    );
    process.exit(1);
  }

  console.log(`[dev-proxy] target=${target} (authenticated)`);
  return createProxyRoutes(target, true);
}

export function resolveDevProxyTarget(): string {
  return process.env.DO_WORKER_API_URL ?? process.env.AGENTCLOUD_API_URL ?? "http://localhost:10000";
}
