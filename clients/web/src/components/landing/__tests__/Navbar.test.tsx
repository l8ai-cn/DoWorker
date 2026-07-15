import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Navbar } from "../Navbar";

const labels = vi.hoisted(() => ({
  "landing.nav.home": "Home",
  "landing.nav.workflow": "How it works",
  "landing.nav.scenarios": "Scenarios",
  "landing.nav.capabilities": "Capabilities",
  "landing.nav.marketplace": "Marketplace",
  "landing.nav.pricing": "Pricing",
  "landing.nav.docs": "Docs",
  "landing.nav.language": "Language",
  "landing.nav.toggleMenu": "Open navigation",
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: keyof typeof labels) => labels[key] ?? key,
}));
vi.mock("next/navigation", () => ({
  usePathname: () => "/capabilities",
}));
vi.mock("@/components/i18n", () => ({
  LanguageSwitcher: () => <button type="button">Language</button>,
}));
vi.mock("@/components/common", () => ({
  LightAuthButtons: () => <div>Auth</div>,
  Logo: () => <span>Logo</span>,
}));

describe("Navbar", () => {
  it("links every primary menu item to its own page", () => {
    render(<Navbar />);

    const expectedLinks = [
      ["Home", "/"],
      ["Scenarios", "/solutions"],
      ["How it works", "/how-it-works"],
      ["Capabilities", "/capabilities"],
      ["Marketplace", "/marketplace"],
      ["Docs", "/docs"],
    ] as const;

    for (const [name, href] of expectedLinks) {
      expect(screen.getAllByRole("link", { name })[0]).toHaveAttribute("href", href);
    }

    expect(screen.getAllByRole("link", { name: "Capabilities" })[0]).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("button", { name: "Open navigation" })).toBeVisible();
    expect(screen.queryByRole("link", { name: "Pricing" })).not.toBeInTheDocument();
  });
});
