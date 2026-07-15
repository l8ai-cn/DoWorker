import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import DocsShell from "../DocsShell";

const labels = vi.hoisted(() => ({
  "docs.nav.menu": "Documentation menu",
  "docs.title": "Documentation",
  "landing.nav.home": "Home",
  "landing.nav.scenarios": "Scenarios",
  "landing.nav.workflow": "How it works",
  "landing.nav.capabilities": "Capabilities",
  "landing.nav.marketplace": "Marketplace",
  "landing.nav.docs": "Docs",
  "common.allRightsReserved": "All rights reserved",
  "landing.footer.legal.privacy": "Privacy",
  "landing.footer.legal.terms": "Terms",
}));

vi.mock("next/navigation", () => ({
  usePathname: () => "/docs",
}));
vi.mock("next-intl", () => ({
  useTranslations: () => (key: keyof typeof labels) => labels[key] ?? key,
}));
vi.mock("@/lib/docs-navigation", () => ({
  getBreadcrumbs: () => [],
}));
vi.mock("@/components/common", () => ({
  LightAuthButtons: () => <div>Auth</div>,
  Logo: () => <span>Logo</span>,
}));
vi.mock("@/components/ui/button", () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button {...props}>{children}</button>
  ),
}));
vi.mock("@/components/ui/sheet", () => ({
  Sheet: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SheetTrigger: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  SheetContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SheetHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SheetTitle: ({ children }: { children: React.ReactNode }) => <h2>{children}</h2>,
}));
vi.mock("../DocsArticle", () => ({
  DocsArticle: ({ children }: { children: React.ReactNode }) => <article>{children}</article>,
}));
vi.mock("../DocsBreadcrumbJsonLd", () => ({
  DocsBreadcrumbJsonLd: () => null,
}));
vi.mock("../DocsSidebarNav", () => ({
  DocsSidebarNav: () => <nav>Documentation sections</nav>,
}));

describe("DocsShell", () => {
  it("keeps every marketing page reachable from the documentation", () => {
    render(<DocsShell>Documentation content</DocsShell>);

    for (const [name, href] of [
      ["Home", "/"],
      ["Scenarios", "/solutions"],
      ["How it works", "/how-it-works"],
      ["Capabilities", "/capabilities"],
      ["Marketplace", "/marketplace"],
      ["Docs", "/docs"],
    ] as const) {
      expect(screen.getAllByRole("link", { name })[0]).toHaveAttribute("href", href);
    }

    expect(screen.getAllByRole("link", { name: "Docs" })[0]).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getAllByRole("link", { name: "Home" })[0]).not.toHaveAttribute(
      "aria-current",
    );
    expect(
      screen
        .getAllByRole("navigation", { name: "Do Worker" })
        .some((navigation) => navigation.className.includes("lg:flex")),
    ).toBe(true);
  });
});
