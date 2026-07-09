import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { mergeMessageNamespaces } from "@/lib/i18n/messageFallback";
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

function readWorkforce(locale: string): MessageTree {
  const path = resolve(__dirname, `../../../messages/${locale}/workforce.json`);
  return JSON.parse(readFileSync(path, "utf8")) as MessageTree;
}

describe("workforce message namespaces", () => {
  it("deep-merges sibling namespaces under landing", () => {
    const landing = { landing: { navbar: { product: "Product" } } };
    const workforce = { landing: { workforce: { hero: { badge: "AI Workforce" } } } };

    expect(mergeMessageNamespaces([landing, workforce])).toEqual({
      landing: {
        navbar: { product: "Product" },
        workforce: { hero: { badge: "AI Workforce" } },
      },
    });
  });

  it("keeps the exact English workforce key set in every locale", () => {
    const englishKeys = flattenKeys(readWorkforce("en")).sort();

    for (const locale of locales) {
      expect(flattenKeys(readWorkforce(locale)).sort(), locale).toEqual(englishKeys);
    }
  });
});
