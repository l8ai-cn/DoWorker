import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { mergeMessageNamespaces } from "@/lib/i18n/messageFallback";
import { locales } from "@/lib/i18n/config";
import { requiredWorkforceMessageKeys } from "@/components/landing/workforce/workforce-message-keys";

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

function channelName(messages: MessageTree): string {
  const landing = messages.landing as MessageTree;
  const workforce = landing.workforce as MessageTree;
  const lifecycle = workforce.lifecycle as MessageTree;
  const fragments = lifecycle.fragments as MessageTree;
  const channel = fragments.channel as MessageTree;
  return channel.name as string;
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

  it("contains exactly the keys required by production workforce components", () => {
    const englishKeys = flattenKeys(readWorkforce("en")).sort();

    expect(englishKeys).toEqual([...requiredWorkforceMessageKeys].sort());
  });

  it("keeps the exact production workforce key set in every locale", () => {
    for (const locale of locales) {
      expect(flattenKeys(readWorkforce(locale)).sort(), locale).toEqual(
        [...requiredWorkforceMessageKeys].sort(),
      );
    }
  });

  it("uses slug-style channel identifiers in every locale", () => {
    for (const locale of locales) {
      expect(channelName(readWorkforce(locale)), locale).toMatch(/^#[a-z0-9]+(?:-[a-z0-9]+)*$/);
    }
  });
});
