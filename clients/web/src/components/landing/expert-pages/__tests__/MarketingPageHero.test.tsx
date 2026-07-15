import { existsSync } from "node:fs";
import { resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const messages = vi.hoisted(() => ({
  "landing.nav.product": "Product",
  "landing.nav.solutions": "Solutions",
  "landing.workforce.expertHome.solutions.title": "Three organization models",
  "landing.workforce.expertHome.solutions.description": "Apply one Agent supply system in three contexts.",
  "landing.workforce.expertHome.hero.primaryAction": "Explore the Agent market",
  "landing.workforce.expertHome.hero.proof": "Shared context and evidence-backed delivery",
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: keyof typeof messages) => messages[key] ?? key,
}));

describe("MarketingPageHero", () => {
  it("introduces an independent content page and its next destination", async () => {
    const componentPath = resolve(__dirname, "../MarketingPageHero.tsx");
    expect(existsSync(componentPath)).toBe(true);
    if (!existsSync(componentPath)) return;

    const componentUrl = pathToFileURL(componentPath).href;
    const { MarketingPageHero } = await import(/* @vite-ignore */ componentUrl);
    render(<MarketingPageHero page="solutions" />);

    expect(screen.getByRole("heading", { level: 1, name: "Three organization models" })).toBeVisible();
    expect(screen.getByText("Apply one Agent supply system in three contexts.")).toBeVisible();
    expect(screen.getByRole("link", { name: "Explore the Agent market" })).toHaveAttribute(
      "href",
      "/marketplace",
    );
    expect(screen.getByRole("link", { name: "Product" })).toHaveAttribute("href", "/product");
  });
});
