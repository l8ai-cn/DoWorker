// @lovable.dev/vite-tanstack-config already includes the following — do NOT add them manually
// or the app will break with duplicate plugins:
//   - tanstackStart, viteReact, tailwindcss, tsConfigPaths, nitro (build-only using cloudflare as a default target),
//     componentTagger (dev-only), VITE_* env injection, @ path alias, React/TanStack dedupe,
//     error logger plugins, and sandbox detection (port/host/strictPort).
// You can pass additional config via defineConfig({ vite: { ... }, etc... }) if needed.
import { defineConfig } from "@lovable.dev/vite-tanstack-config";
import type { ProxyOptions } from "vite";

const devBackendUrl =
  process.env.DO_WORKER_API_URL ?? process.env.VITE_DO_WORKER_API_URL ?? "http://127.0.0.1:10015";

export default defineConfig({
  nitro: { preset: "node-server" },
  tanstackStart: {
    server: { entry: "server" },
  },
  vite: {
    server: {
      proxy: {
        "/auth": proxyTarget(devBackendUrl),
        "/v1": proxyTarget(devBackendUrl),
        "/api": proxyTarget(devBackendUrl),
        "/proto": proxyTarget(devBackendUrl),
      },
    },
  },
});

function proxyTarget(target: string): ProxyOptions {
  return {
    target,
    changeOrigin: true,
    ws: true,
    configure: (proxy) => {
      proxy.on("proxyReq", (proxyReq) => {
        proxyReq.removeHeader("origin");
      });
    },
  };
}
