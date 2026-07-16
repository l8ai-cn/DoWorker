import { describe, expect, it } from "vitest";
import workerRuntimeCatalog from "@/generated/worker-runtime-catalog.json";
import enExpertHome from "@/messages/en/expert-home.json";
import enLanding from "@/messages/en/landing.json";
import zhExpertHome from "@/messages/zh/expert-home.json";
import zhLanding from "@/messages/zh/landing.json";

import {
  capabilityGroups,
  marketplaceApplications,
  solutionDomains,
  workerTypes,
} from "../expert-home-content";

const solutionIds = [
  "enterprise-agent-supply",
  "opc-incubation",
  "higher-education-digital-employees",
] as const;

const capabilityIds = [
  "agent-factory",
  "agent-market",
  "collaboration-workspace",
  "automation",
  "governance",
] as const;

const lifecycleIds = ["build", "verify", "release", "install", "run", "evolve"] as const;

const invocationSteps = {
  en: [
    "Requirement identification",
    "Capability matching",
    "Permission confirmation",
    "Collaborative execution",
    "Human review",
    "Evidence delivery",
  ],
  zh: ["需求识别", "能力匹配", "权限确认", "协同执行", "人工复核", "证据交付"],
} as const;

const localizedContent = {
  en: enExpertHome.landing.workforce.expertHome,
  zh: zhExpertHome.landing.workforce.expertHome,
};

function collectMarketingCopy(value: unknown): string[] {
  if (typeof value === "string") return [value];
  if (Array.isArray(value)) return value.flatMap(collectMarketingCopy);
  if (value && typeof value === "object") {
    return Object.entries(value).flatMap(([key, child]) =>
      ["id", "slug", "level"].includes(key) ? [] : collectMarketingCopy(child),
    );
  }
  return [];
}

describe("expert homepage content contract", () => {
  it("exposes exactly the three approved solution domains", () => {
    expect(solutionDomains.map(({ id }) => id)).toEqual(solutionIds);
  });

  it("lists every formal worker type from the product catalog", () => {
    expect(workerTypes).toEqual(
      workerRuntimeCatalog.workers.map(({ slug, name }) => ({ slug, name })),
    );
  });

  it("keeps the Agent market aligned with published Agent applications", () => {
    expect(marketplaceApplications.map(({ slug }) => slug)).toEqual([
      "software-delivery-expert",
      "multi-worker-orchestrator",
      "dual-repo-sync-expert",
    ]);
  });

  it("defines the five supply-first platform capabilities", () => {
    expect(capabilityGroups).toEqual(
      capabilityIds.map((id) => ({ id, level: "implemented" })),
    );
  });

  it("keeps English and Chinese arrays aligned with the product contract", () => {
    for (const [locale, content] of Object.entries(localizedContent)) {
      expect(content.solutions.items.map(({ id }) => id), locale).toEqual(solutionIds);
      expect(content.capabilities.items.map(({ id }) => id), locale).toEqual(capabilityIds);
      expect(content.operating.parts.map(({ id }) => id), locale).toEqual(lifecycleIds);
      expect(content.market.apps.map(({ slug }) => slug), locale).toEqual(
        marketplaceApplications.map(({ slug }) => slug),
      );
      expect(content.console.steps, locale).toEqual(
        invocationSteps[locale as keyof typeof invocationSteps],
      );
    }
  });

  it("leads with Agent supply and AI-native organization incubation", () => {
    expect(localizedContent.en.hero.title).toBe(
      "Agent supply + AI-native organization incubation",
    );
    expect(localizedContent.zh.hero.title).toBe("Agent 供给 + AI 原生组织孵化");
    expect(localizedContent.en.hero.primaryAction).toBe("Explore the Agent market");
    expect(localizedContent.en.hero.secondaryAction).toBe("Explore the supply system");
    expect(localizedContent.zh.hero.primaryAction).toBe("进入 Agent 市场");
    expect(localizedContent.zh.hero.secondaryAction).toBe("了解供给体系");
    expect(localizedContent.en.cta.primary).toBe("Build your first Agent");
    expect(localizedContent.zh.cta.primary).toBe("构建第一个 Agent");

    for (const [locale, content] of Object.entries(localizedContent)) {
      expect(collectMarketingCopy(content).join("\n"), locale).not.toMatch(/\bexperts?\b/i);
    }
  });

  it("positions higher-education digital employees as a pilot direction", () => {
    expect(localizedContent.en.solutions.items[2]).toMatchObject({
      id: "higher-education-digital-employees",
      action: "Start a higher-education pilot",
    });
    expect(localizedContent.zh.solutions.items[2]).toMatchObject({
      id: "higher-education-digital-employees",
      action: "启动高校数字员工试点",
    });
    expect(localizedContent.en.solutions.items[2].description).toContain(
      "composable platform capabilities",
    );
    expect(localizedContent.zh.solutions.items[2].description).toContain(
      "可组合的平台能力",
    );
  });

  it("keeps technical implementation terms out of customer-facing content", () => {
    for (const [locale, content] of Object.entries(localizedContent)) {
      const { trust: governance, ...nonGovernanceContent } = content;
      expect(collectMarketingCopy(content).join("\n"), locale).not.toMatch(
        /\b(?:Pod|AgentPod|WorkerSpec|ResourceRef)\b/i,
      );
      expect(collectMarketingCopy(nonGovernanceContent).join("\n"), locale).not.toMatch(
        /\b(?:Worker|Runner)\b/i,
      );
      expect(collectMarketingCopy(content.cta).join("\n"), locale).not.toMatch(
        /\b(?:free|pricing)\b/i,
      );
      expect(collectMarketingCopy(governance).length, locale).toBeGreaterThan(0);
    }
  });

  it("adds product and solutions navigation without removing existing entries", () => {
    expect(enLanding.landing.nav).toMatchObject({
      product: "Product",
      solutions: "Solutions",
      marketplace: "Agent Market",
    });
    expect(zhLanding.landing.nav).toMatchObject({
      product: "产品",
      solutions: "解决方案",
      marketplace: "Agent 市场",
    });
  });

  it("keeps the global marketing tagline supply-first", () => {
    expect(enLanding.landing.footer.tagline).toBe(
      "Supply Agents. Incubate AI-native organizations.",
    );
    expect(zhLanding.landing.footer.tagline).toBe("供给 Agent，孵化 AI 原生组织。");
  });
});
