import { existsSync } from "node:fs";
import { resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const messages = vi.hoisted(() => ({
  "landing.nav.scenarios": "Scenarios",
  "landing.nav.workflow": "How it works",
  "landing.workforce.expertHome.solutions.title": "Outcome-first Experts",
  "landing.workforce.expertHome.solutions.description": "Carry one business goal to delivery.",
  "landing.workforce.expertHome.hero.primaryAction": "Create your Expert",
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

    expect(screen.getByRole("heading", { level: 1, name: "Outcome-first Experts" })).toBeVisible();
    expect(screen.getByText("Carry one business goal to delivery.")).toBeVisible();
    expect(screen.getByRole("link", { name: "Create your Expert" })).toHaveAttribute("href", "/register");
    expect(screen.getByRole("link", { name: "How it works" })).toHaveAttribute("href", "/how-it-works");
  });
});
