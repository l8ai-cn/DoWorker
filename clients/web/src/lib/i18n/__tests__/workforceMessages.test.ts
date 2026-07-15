import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { mergeMessageNamespaces } from "@/lib/i18n/messageFallback";
import { locales } from "@/lib/i18n/config";
import { requiredExpertHomeMessageKeys } from "@/components/landing/expert-home/expert-home-message-keys";
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
  const workforcePath = resolve(__dirname, `../../../messages/${locale}/workforce.json`);
  const expertHomePath = resolve(__dirname, `../../../messages/${locale}/expert-home.json`);
  return mergeMessageNamespaces([
    JSON.parse(readFileSync(workforcePath, "utf8")) as MessageTree,
    JSON.parse(readFileSync(expertHomePath, "utf8")) as MessageTree,
  ]);
}

function channelName(messages: MessageTree): string {
  const landing = messages.landing as MessageTree;
  const workforce = landing.workforce as MessageTree;
  const lifecycle = workforce.lifecycle as MessageTree;
  const fragments = lifecycle.fragments as MessageTree;
  const channel = fragments.channel as MessageTree;
  return channel.name as string;
}

function aiPartnerTitle(messages: MessageTree): string {
  const landing = messages.landing as MessageTree;
  const workforce = landing.workforce as MessageTree;
  const expertHome = workforce.expertHome as MessageTree;
  const solutions = expertHome.solutions as MessageTree;
  const items = solutions.items as MessageTree[];
  const aiPartner = items.find(({ id }) => id === "digital-employees");
  return aiPartner?.title as string;
}

describe("workforce message namespaces", () => {
  const requiredKeys = [
    ...requiredWorkforceMessageKeys,
    ...requiredExpertHomeMessageKeys,
  ];

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

    expect(englishKeys).toEqual([...requiredKeys].sort());
  });

  it("keeps the exact production workforce key set in every locale", () => {
    for (const locale of locales) {
      expect(flattenKeys(readWorkforce(locale)).sort(), locale).toEqual(
        [...requiredKeys].sort(),
      );
    }
  });

  it("uses slug-style channel identifiers in every locale", () => {
    for (const locale of locales) {
      expect(channelName(readWorkforce(locale)), locale).toMatch(/^#[a-z0-9]+(?:-[a-z0-9]+)*$/);
    }
  });

  it("names the recurring-work solution as AI partners in every locale", () => {
    const expectedTitles: Record<string, string> = {
      de: "KI-Partner",
      en: "AI partners",
      es: "Compañeros de IA",
      fr: "Partenaires IA",
      ja: "AIパートナー",
      ko: "AI 파트너",
      pt: "Parceiros de IA",
      zh: "AI 伙伴",
    };

    for (const locale of locales) {
      expect(aiPartnerTitle(readWorkforce(locale)), locale).toBe(expectedTitles[locale]);
    }
  });
});
