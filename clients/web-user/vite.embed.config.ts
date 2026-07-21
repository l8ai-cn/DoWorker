import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

import { resolveExternalCjsRequire } from "./vite-embed-cjs-externals";
import { scopeAgentCloudCss } from "./vite-embed-css-scope";

const sharedExternals = [
  "react",
  "react-dom",
  "react/jsx-runtime",
  "react-router",
  "react-router-dom",
];

export default defineConfig({
  base: "./",
  plugins: [react(), tailwindcss(), scopeAgentCloudCss(), resolveExternalCjsRequire(sharedExternals)],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@agent-cloud/agent-ui": path.resolve(__dirname, "../../packages/agent-ui/src/index.ts"),
      "react-markdown": path.resolve(__dirname, "./node_modules/react-markdown/index.js"),
      "remark-gfm": path.resolve(__dirname, "./node_modules/remark-gfm/index.js"),
    },
    dedupe: ["lucide-react", "react", "react-dom"],
  },
  define: {
    "process.env.NODE_ENV": JSON.stringify("production"),
  },
  build: {
    outDir: path.resolve(__dirname, "./dist-embed"),
    emptyOutDir: true,
    cssCodeSplit: false,
    sourcemap: false,
    lib: {
      entry: path.resolve(__dirname, "./src/embed.tsx"),
      formats: ["es"],
    },
    rollupOptions: {
      external: sharedExternals,
      output: {
        entryFileNames: "agent-cloud-embed.js",
        chunkFileNames: "chunks/[name]-[hash].js",
        assetFileNames: (assetInfo) => {
          const name = assetInfo.names?.[0] ?? "";
          return name.endsWith(".css") ? "agent-cloud-embed.css" : "assets/[name].[ext]";
        },
      },
    },
  },
});
