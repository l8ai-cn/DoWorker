import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { locales } from "@/lib/i18n/config";

type MessageTree = Record<string, unknown>;

function flattenKeys(value: MessageTree, prefix = ""): string[] {
  return Object.entries(value).flatMap(([key, child]) => {
    const path = prefix ? `${prefix}.${key}` : key;
    return typeof child === "object" && child !== null && !Array.isArray(child)
      ? flattenKeys(child as MessageTree, path)
      : [path];
  });
}

function readMessages(locale: string): MessageTree {
  const path = resolve(
    __dirname,
    `../../../messages/${locale}/resource-orchestration.json`,
  );
  return JSON.parse(readFileSync(path, "utf8")) as MessageTree;
}

describe("resource orchestration messages", () => {
  it("keeps the exact English key set in every locale", () => {
    const expected = flattenKeys(readMessages("en")).sort();

    for (const locale of locales) {
      expect(flattenKeys(readMessages(locale)).sort(), locale).toEqual(expected);
    }
  });
});
