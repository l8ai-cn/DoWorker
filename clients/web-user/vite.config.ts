import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

import { createDevProxyConfig, resolveDevProxyTarget } from "./vite-dev-proxy";
import { emitPwaAssets } from "./vite-pwa-assets";

const DO_WORKER_API_URL = resolveDevProxyTarget();
const proxyConfig = createDevProxyConfig(DO_WORKER_API_URL);

export default defineConfig({
  plugins: [emitPwaAssets(__dirname), react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@do-worker/agent-ui": path.resolve(__dirname, "../../packages/agent-ui/src/index.ts"),
      "react-markdown": path.resolve(__dirname, "./node_modules/react-markdown/index.js"),
      "remark-gfm": path.resolve(__dirname, "./node_modules/remark-gfm/index.js"),
    },
    dedupe: ["lucide-react", "react", "react-dom"],
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./src/test-setup.ts"],
    // Scope discovery to src/ — the web suite lives there. Without this,
    // vitest's default glob descends into the nested electron package and
    // tries to run its node:test files (which aren't vitest suites).
    include: ["src/**/*.{test,spec}.?(c|m)[jt]s?(x)"],
    coverage: {
      provider: "v8",
      // With `include` set, vitest counts every matching source file (untested
      // ones as 0%), so the total reflects the whole frontend — parity with the
      // backend's coverage scope, not just files a test happened to import.
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.test.{ts,tsx}",
        "src/**/*.d.ts",
        "src/test-setup.ts",
        // Vendored UI kit, not product code (see tests/e2e_ui/COVERAGE_GAPS.md).
        "src/components/ai-elements/**",
      ],
      reportsDirectory: "./coverage",
      // text-summary: human-readable console line; json-summary: machine-
      // readable coverage/coverage-summary.json that CI distills to total.txt.
      reporter: ["text-summary", "json-summary"],
    },
  },
  server: {
    proxy: proxyConfig,
  },
  build: {
    outDir: path.resolve(__dirname, "./dist"),
    emptyOutDir: true,
    rollupOptions: {
      input: {
        index: path.resolve(__dirname, "./index.html"),
        iframe: path.resolve(__dirname, "./iframe.html"),
        previewWindow: path.resolve(__dirname, "./preview-window.html"),
        worker: path.resolve(__dirname, "./worker.html"),
      },
    },
  },
});
