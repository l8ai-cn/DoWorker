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

function solutionTitle(messages: MessageTree, solutionId: string): string {
  const landing = messages.landing as MessageTree;
  const workforce = landing.workforce as MessageTree;
  const expertHome = workforce.expertHome as MessageTree;
  const solutions = expertHome.solutions as MessageTree;
  const items = solutions.items as MessageTree[];
  const solution = items.find(({ id }) => id === solutionId);
  return solution?.title as string;
}

function contentIds(
  messages: MessageTree,
  sectionName: "solutions" | "capabilities" | "operating",
  itemName: "items" | "parts",
): string[] {
  const landing = messages.landing as MessageTree;
  const workforce = landing.workforce as MessageTree;
  const expertHome = workforce.expertHome as MessageTree;
  const section = expertHome[sectionName] as MessageTree;
  const items = section[itemName] as MessageTree[];
  return items.map(({ id }) => id as string);
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

  it("uses the approved supply-first solution names in English and Chinese", () => {
    const expectedTitles = {
      en: {
        "enterprise-agent-supply": "Enterprise Agent supply",
        "opc-incubation": "OPC incubation",
        "higher-education-digital-employees": "Higher-education digital employees",
      },
      zh: {
        "enterprise-agent-supply": "企业 Agent 供给",
        "opc-incubation": "OPC 孵化",
        "higher-education-digital-employees": "高校数字员工",
      },
    };

    for (const [locale, titles] of Object.entries(expectedTitles)) {
      const messages = readWorkforce(locale);
      for (const [solutionId, title] of Object.entries(titles)) {
        expect(solutionTitle(messages, solutionId), `${locale}:${solutionId}`).toBe(title);
      }
    }
  });

  it("keeps every locale on the same supply-first content structure", () => {
    for (const locale of locales) {
      const messages = readWorkforce(locale);
      expect(contentIds(messages, "solutions", "items"), locale).toEqual([
        "enterprise-agent-supply",
        "opc-incubation",
        "higher-education-digital-employees",
      ]);
      expect(contentIds(messages, "capabilities", "items"), locale).toEqual([
        "agent-factory",
        "agent-market",
        "collaboration-workspace",
        "automation",
        "governance",
      ]);
      expect(contentIds(messages, "operating", "parts"), locale).toEqual([
        "build",
        "verify",
        "release",
        "install",
        "run",
        "evolve",
      ]);
    }
  });
});
