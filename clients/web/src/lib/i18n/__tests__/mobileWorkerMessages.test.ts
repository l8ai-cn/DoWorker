import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
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

function readMobileMessages(locale: string): MessageTree {
  const path = resolve(__dirname, `../../../messages/${locale}/app.json`);
  const app = JSON.parse(readFileSync(path, "utf8")) as MessageTree;
  const mobile = app.mobile as MessageTree;
  return {
    access: mobile.access,
    control: mobile.control,
    preview: mobile.preview,
    workers: mobile.workers,
  };
}

describe("mobile Worker messages", () => {
  it("keeps every mobile Worker key in every locale", () => {
    const expected = flattenKeys(readMobileMessages("en")).sort();

    for (const locale of locales) {
      expect(flattenKeys(readMobileMessages(locale)).sort(), locale).toEqual(
        expected,
      );
    }
  });
});
