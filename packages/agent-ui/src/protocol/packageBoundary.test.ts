import { readFileSync, readdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const sourceRoot = dirname(dirname(fileURLToPath(import.meta.url)));

describe("agent-ui protocol package boundary", () => {
  it("uses an explicit proto package instead of imports outside agent-ui", () => {
    const sourceFiles = ["protocol", "runtime"].flatMap((directory) =>
      readdirSync(join(sourceRoot, directory), {
        recursive: true,
        withFileTypes: true,
      })
        .filter((entry) => entry.isFile() && entry.name.endsWith(".ts"))
        .map((entry) => join(entry.parentPath, entry.name))
        .filter((path) => !path.endsWith(".test.ts")),
    );

    for (const sourceFile of sourceFiles) {
      expect(readFileSync(sourceFile, "utf8")).not.toContain("proto/gen/ts");
    }

    const packageJson = JSON.parse(
      readFileSync(join(sourceRoot, "..", "package.json"), "utf8"),
    ) as { dependencies?: Record<string, string> };
    expect(packageJson.dependencies?.["@do-worker/proto"]).toBe("workspace:*");
  });
});
