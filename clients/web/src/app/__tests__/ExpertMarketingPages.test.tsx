import { existsSync } from "node:fs";
import { resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("@/components/landing/expert-pages/MarketingPageShell", () => ({
  MarketingPageShell: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="marketing-shell">{children}</div>
  ),
}));
vi.mock("@/components/landing/expert-pages/MarketingPageHero", () => ({
  MarketingPageHero: ({ page }: { page: string }) => <div data-testid={`hero-${page}`} />,
}));
vi.mock("@/components/landing/expert-home/SolutionDomains", () => ({
  SolutionDomains: ({ showIntro }: { showIntro?: boolean }) => (
    <div data-testid="solutions-content" data-show-intro={showIntro} />
  ),
}));
vi.mock("@/components/landing/expert-home/ExpertOperatingModel", () => ({
  ExpertOperatingModel: ({ showIntro }: { showIntro?: boolean }) => (
    <div data-testid="operating-content" data-show-intro={showIntro} />
  ),
}));
vi.mock("@/components/landing/expert-home/CapabilitySpectrum", () => ({
  CapabilitySpectrum: ({ showIntro }: { showIntro?: boolean }) => (
    <div data-testid="capabilities-content" data-show-intro={showIntro} />
  ),
}));
vi.mock("@/components/landing/expert-home/ExpertGovernance", () => ({
  ExpertGovernance: () => <div data-testid="governance-content" />,
}));

afterEach(cleanup);

const cases = [
  {
    route: "solutions",
    hero: "hero-solutions",
    regions: ["solutions-content"],
  },
  {
    route: "product",
    hero: "hero-product",
    regions: ["operating-content", "capabilities-content", "governance-content"],
  },
] as const;

describe("Agent supply marketing pages", () => {
  for (const pageCase of cases) {
    it(`composes the ${pageCase.route} page from its dedicated regions`, async () => {
      const pagePath = resolve(__dirname, `../${pageCase.route}/page.tsx`);
      expect(existsSync(pagePath)).toBe(true);
      if (!existsSync(pagePath)) return;

      const pageUrl = pathToFileURL(pagePath).href;
      const { default: Page } = await import(/* @vite-ignore */ pageUrl);
      render(<Page />);

      expect(screen.getByTestId("marketing-shell")).toBeVisible();
      expect(screen.getByTestId(pageCase.hero)).toBeVisible();
      for (const region of pageCase.regions) {
        const content = screen.getByTestId(region);
        expect(content).toBeVisible();
        if (region !== "governance-content") {
          expect(content).toHaveAttribute("data-show-intro", "false");
        }
      }
    });
  }
});
