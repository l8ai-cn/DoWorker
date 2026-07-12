import react from "@vitejs/plugin-react";
import { resolve } from "node:path";
import { defineConfig } from "vitest/config";

const projectRoot = __dirname;

export default defineConfig({
  root: projectRoot,
  plugins: [react()],
  resolve: {
    alias: {
      "@": resolve(projectRoot, "./src"),
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: [resolve(projectRoot, "./src/test/setup.ts")],
  },
});
