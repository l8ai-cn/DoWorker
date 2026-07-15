import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Navbar } from "../Navbar";

const labels = vi.hoisted(() => ({
  "landing.nav.home": "Home",
  "landing.nav.product": "Product",
  "landing.nav.solutions": "Solutions",
  "landing.nav.marketplace": "Agent Market",
  "landing.nav.docs": "Docs",
  "landing.nav.language": "Language",
  "landing.nav.toggleMenu": "Open navigation",
}));

vi.mock("next-intl", () => ({
  useTranslations: () => (key: keyof typeof labels) => labels[key] ?? key,
}));
vi.mock("next/navigation", () => ({
  usePathname: () => "/product",
}));
vi.mock("@/components/i18n", () => ({
  LanguageSwitcher: () => <button type="button">Language</button>,
}));
vi.mock("@/components/common/LightAuthButtons", () => ({
  LightAuthButtons: () => <div>Auth</div>,
}));
vi.mock("@/components/common/Logo", () => ({
  Logo: () => <span>Logo</span>,
}));

describe("Navbar", () => {
  it("links every primary menu item to its own page", () => {
    render(<Navbar />);

    const expectedLinks = [
      ["Home", "/"],
      ["Product", "/product"],
      ["Solutions", "/solutions"],
      ["Agent Market", "/marketplace"],
      ["Docs", "/docs"],
    ] as const;

    for (const [name, href] of expectedLinks) {
      expect(screen.getAllByRole("link", { name })[0]).toHaveAttribute("href", href);
    }

    expect(screen.getAllByRole("link", { name: "Product" })[0]).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("button", { name: "Open navigation" })).toBeVisible();
  });
});
