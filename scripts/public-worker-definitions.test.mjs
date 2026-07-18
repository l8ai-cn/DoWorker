import assert from "node:assert/strict";
import test from "node:test";
import { loadPublicWorkerDefinitions } from "./public-worker-definitions.mjs";

test("excludes internal definitions from public catalogs", () => {
  const definitions = new Map([
    ["/repo/config/public.json", { internal: false }],
    ["/repo/config/internal.json", { internal: true }],
  ]);

  const result = loadPublicWorkerDefinitions({
    definitionCatalog: {
      worker_types: [
        { slug: "public-worker", definition_path: "config/public.json" },
        { slug: "internal-worker", definition_path: "config/internal.json" },
      ],
    },
    readJson: (filePath) => definitions.get(filePath),
    root: "/repo",
  });

  assert.deepEqual(
    result.map(({ entry }) => entry.slug),
    ["public-worker"],
  );
});
