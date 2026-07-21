import { readFileSync } from "node:fs";
import path from "node:path";
import type { Plugin } from "vite";

import { computeBuildVersion } from "./src/lib/buildVersion";

const PWA_MANIFEST = {
  id: "/",
  name: "Agent Cloud",
  short_name: "Agent Cloud",
  description: "Agent Cloud — a common layer over coding agents.",
  start_url: "/",
  scope: "/",
  display: "standalone",
  orientation: "any",
  theme_color: "#0d1218",
  background_color: "#0d1218",
  icons: [
    { src: "/pwa-192.png", sizes: "192x192", type: "image/png" },
    { src: "/pwa-512.png", sizes: "512x512", type: "image/png" },
    { src: "/pwa-maskable-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
  ],
};

export function emitPwaAssets(projectDir: string): Plugin {
  return {
    name: "emit-pwa-assets",
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (req.url !== "/manifest.webmanifest") return next();
        res.setHeader("Content-Type", "application/manifest+json");
        res.end(JSON.stringify(PWA_MANIFEST));
      });
    },
    generateBundle(_options, bundle) {
      const build = computeBuildVersion(Object.keys(bundle));
      const swSource = readFileSync(path.resolve(projectDir, "sw-src/sw.js"), "utf8");
      if (!swSource.includes("__BUILD_VERSION__")) {
        this.error("sw-src/sw.js is missing the __BUILD_VERSION__ token; cannot fingerprint sw.js");
      }

      this.emitFile({
        type: "asset",
        fileName: "version.json",
        source: JSON.stringify({ build }),
      });
      this.emitFile({
        type: "asset",
        fileName: "manifest.webmanifest",
        source: JSON.stringify(PWA_MANIFEST),
      });
      this.emitFile({
        type: "asset",
        fileName: "sw.js",
        source: swSource.replaceAll("__BUILD_VERSION__", build),
      });
    },
  };
}
