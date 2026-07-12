// @lovable.dev/vite-tanstack-config already includes the following — do NOT add them manually
// or the app will break with duplicate plugins:
//   - tanstackStart, viteReact, tailwindcss, tsConfigPaths, nitro (build-only using cloudflare as a default target),
//     componentTagger (dev-only), VITE_* env injection, @ path alias, React/TanStack dedupe,
//     error logger plugins, and sandbox detection (port/host/strictPort).
// You can pass additional config via defineConfig({ vite: { ... }, etc... }) if needed.
import { defineConfig } from "@lovable.dev/vite-tanstack-config";
import type { ProxyOptions } from "vite";

export default defineConfig({
  nitro: { preset: "node-server" },
  tanstackStart: {
    server: { entry: "server" },
  },
  vite: {
    server: {
      proxy: {
        "/auth": proxyTarget("http://localhost:10000"),
        "/v1": proxyTarget("http://localhost:10000"),
        "/api": proxyTarget("http://localhost:10000"),
        "/proto": proxyTarget("http://localhost:10000"),
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
